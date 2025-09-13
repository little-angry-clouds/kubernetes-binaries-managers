package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/mholt/archiver/v3"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	goVersion "go.hein.dev/go-version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// =============================================================================
// CONFIGURATION - Embedded tool configurations
// =============================================================================

// ToolConfig defines configuration for each supported tool
type ToolConfig struct {
	Name           string
	DownloadURL    string
	VersionsAPI    string
	Description    string
	SupportsAuto   bool
	CompressedType string // "binary", "tar.gz", "zip"
}

// supportedTools is the registry of all supported tools with their configurations
var supportedTools = map[string]ToolConfig{
	"kubectl": {
		Name:           "kubectl",
		DownloadURL:    "https://storage.googleapis.com/kubernetes-release/release/v%s/bin/%s/%s/kubectl",
		VersionsAPI:    "https://api.github.com/repos/kubernetes/kubernetes/releases?per_page=100&page=",
		Description:    "Kubernetes CLI tool",
		SupportsAuto:   true,
		CompressedType: "binary",
	},
	"helm": {
		Name:           "helm",
		DownloadURL:    "https://get.helm.sh/helm-v%s-%s-%s",
		VersionsAPI:    "https://api.github.com/repos/helm/helm/releases?per_page=100&page=",
		Description:    "Kubernetes package manager",
		SupportsAuto:   false,
		CompressedType: "tar.gz",
	},
	"oc": {
		Name:           "oc",
		DownloadURL:    "https://github.com/openshift/okd/releases/download/%s/openshift-client-%s-%s",
		VersionsAPI:    "https://api.github.com/repos/openshift/okd/releases?per_page=100&page=",
		Description:    "OpenShift CLI tool",
		SupportsAuto:   false,
		CompressedType: "tar.gz",
	},
}

// getTool returns tool configuration for given name
func getTool(name string) (ToolConfig, bool) {
	tool, exists := supportedTools[name]
	return tool, exists
}

// getAllTools returns list of all supported tool names
func getAllTools() []string {
	tools := make([]string, 0, len(supportedTools))
	for name := range supportedTools {
		tools = append(tools, name)
	}
	return tools
}

// validateTool checks if tool name is supported
func validateTool(name string) error {
	if _, exists := supportedTools[name]; !exists {
		return fmt.Errorf("unsupported tool: %s. Supported tools: %v", name, getAllTools())
	}
	return nil
}

// =============================================================================
// CONSTANTS AND TYPES
// =============================================================================

const (
	zip   = ".zip"
	targz = ".tar.gz"
	exe   = ".exe"
)

// Version build information (set by build flags)
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Error types
type DownloadBinaryError struct {
	Err  string
	URL  string
	Body string
}

func (e *DownloadBinaryError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("%s\nurl: %s", e.Err, e.URL)
	}
	return fmt.Sprintf("%s\nurl: %s\nbody: %s", e.Err, e.URL, e.Body)
}

type OSArchError struct {
	Err  string
	OS   string
	Arch string
}

func (e *OSArchError) Error() string {
	if e.Arch == "" {
		return fmt.Sprintf("%s\nos: %s", e.Err, e.OS)
	}
	return fmt.Sprintf("%s\narch: %s", e.Err, e.Arch)
}

// Page represents a GitHub release page response
type Page struct {
	Release string `json:"tag_name"`
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// checkGenericError checks if there's an error, shows it and exits the program if it is
func checkGenericError(err error) {
	if err != nil {
		message := fmt.Sprintf("An error was detected, exiting: %s", err)
		fmt.Fprintf(os.Stderr, "%s\n", message)
		os.Exit(1)
	}
}

func checkHTTPError(resp *http.Response) {
	var result map[string]interface{}
	var message string

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if resp.Header.Get("Content-Type") == "application/json" {
			checkGenericError(err)
			err = json.Unmarshal(body, &result)
			checkGenericError(err)
			message = result["message"].(string)
		} else {
			message = string(body)
		}
		fmt.Fprintf(os.Stderr, "An error detected getting all versions: %s", message)
		os.Exit(1)
	}
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getLastPage(link string) (int, error) {
	var lastPageInt int
	var err error
	var minLastPage int = 2

	// When there's no link, it means there's least than 100 releases
	if link == "" {
		return minLastPage, nil
	}

	link = strings.Split(link, " ")[2]
	lastPageIndex := strings.LastIndex(link, "page=")
	lastPageStr := strings.Replace(link[lastPageIndex+5:], ">;", "", 2)
	lastPageInt, err = strconv.Atoi(lastPageStr)

	if err != nil {
		return 0, err
	}

	if lastPageInt == 0 {
		lastPageInt = minLastPage
	}

	return lastPageInt, nil
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func getOSArch() (string, error) {
	var supportedOS = []string{"linux", "windows", "darwin"}
	var supportedArch = []string{"arm", "arm64", "amd64"}
	var os = runtime.GOOS
	var arch = runtime.GOARCH

	if !contains(supportedOS, os) {
		return "", &OSArchError{"os not supported", os, ""}
	}

	if !contains(supportedArch, arch) {
		return "", &OSArchError{"arch not supported", "", arch}
	}

	return os + "/" + arch, nil
}

// =============================================================================
// KUBERNETES CLIENT HELPERS
// =============================================================================

func kubeGetVersion() (string, error) {
	var kubeconfig string
	var cli *string
	var err error
	var home string
	var version string
	var config *rest.Config

	cli = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	// Try first --kubeconfig parameter
	if *cli != "" {
		kubeconfig = *cli
	} else {
		home, _ = homedir.Dir()
		kubeconfigVar := os.Getenv("KUBECONFIG")

		// Try first KUBECONFIG env var
		// If not set, try default kubeconfig path
		// If not find, let it empty
		if kubeconfigVar != "" {
			kubeconfig = kubeconfigVar
		} else if fileExists(filepath.Join(home, ".kube", "config")) {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// If no kubeconfig
	if kubeconfig == "" {
		version = getDefaultVersion()
		return version, nil
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println(fmt.Errorf("binarius failed to load kubeconfig: %w", err))
		os.Exit(1)
	}

	config.Timeout = 1 * time.Second

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		version = getDefaultVersion()
		return version, nil
	}

	v, err := client.DiscoveryClient.ServerVersion()
	if err != nil {
		version = getDefaultVersion()
		return version, nil
	}

	version = v.String()[1:]
	return version, nil
}

func getDefaultVersion() string {
	var fileExt string
	var version string

	if runtime.GOOS == "windows" {
		fileExt = ".exe"
	}

	// See if there's any kubectl version installed
	args := []string{"kubectl", "list", "local"}
	cmd := exec.Command("binarius"+fileExt, args...)
	output, _ := cmd.Output()
	out := string(output)

	// If there's no kubectl, get latest
	// If there's at least one, use that
	if out == "" {
		resp, err := http.Get("https://storage.googleapis.com/kubernetes-release/release/stable.txt")
		checkGenericError(err)

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		checkGenericError(err)

		bodyText := string(body)
		version = strings.Trim(bodyText[1:], "\n")
	} else {
		lines := strings.Split(out, "\n")
		version = lines[0]
	}

	return version
}

// =============================================================================
// BINARY MANAGEMENT
// =============================================================================

// downloadBinaryWithConfig downloads a binary using tool configuration
func downloadBinaryWithConfig(version string, toolConfig ToolConfig) ([]byte, error) {
	var osArch string
	var err error
	var body []byte
	var errorCodeFail = 404
	var errorCodePass = 200

	osArch, _ = getOSArch()
	os := strings.Split(osArch, "/")[0]
	arch := strings.Split(osArch, "/")[1]

	url := toolConfig.DownloadURL

	// Handle tool-specific URL formatting
	switch toolConfig.Name {
	case "oc":
		// They use mac instead of darwin in the url
		if os == "darwin" {
			os = "mac"
		} else if os == "windows" {
			url = strings.Replace(url, ".tar.gz", ".zip", 1)
		}
		url = fmt.Sprintf(url, version, os, version)
	default:
		url = fmt.Sprintf(url, version, os, arch)
	}

	// Add file extensions based on compression type and OS
	switch toolConfig.CompressedType {
	case "tar.gz":
		if strings.Contains(osArch, "windows") {
			url += zip
		} else {
			url += targz
		}
	case "binary":
		if strings.Contains(osArch, "windows") {
			url += exe
		}
		// No extension for binary on unix systems
	case "zip":
		url += zip
	}

	fmt.Println("Downloading binary...")
	resp, err := http.Get(url)
	if err != nil {
		return body, err
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}

	if resp.StatusCode == errorCodeFail {
		return body, &DownloadBinaryError{"binary not found", url, string(body)}
	} else if resp.StatusCode != errorCodePass {
		return body, &DownloadBinaryError{"unhandled error", url, string(body)}
	}

	return body, nil
}

// saveBinaryWithConfig saves a binary using tool configuration
func saveBinaryWithConfig(fileName string, body []byte, toolConfig ToolConfig) error {
	var err error

	// Handle compressed files based on tool configuration
	if toolConfig.CompressedType == "tar.gz" {
		var fileExt string

		rand.Seed(time.Now().UnixNano())

		randomNumbers := 5000
		tempDir, err := ioutil.TempDir("", toolConfig.Name)
		if err != nil {
			return err
		}
		// clean temp dir
		defer os.RemoveAll(tempDir)

		osArch, _ := getOSArch()
		file := fmt.Sprintf("%s/%s-%s", tempDir, toolConfig.Name, strconv.Itoa(rand.Intn(randomNumbers)))
		file, _ = filepath.Abs(file)

		if strings.Contains(osArch, "windows") {
			fileExt = zip
		} else {
			fileExt = targz
		}

		err = ioutil.WriteFile(file+fileExt, body, 0750)
		if err != nil {
			return err
		}

		err = archiver.Unarchive(file+fileExt, file)
		if err != nil {
			return err
		}

		// Extract the binary based on tool type
		var extractPath string
		switch toolConfig.Name {
		case "helm":
			OS := strings.Split(osArch, "/")[0]
			arch := strings.Split(osArch, "/")[1]
			extractPath, _ = filepath.Abs(file + fmt.Sprintf("/%s-%s/helm", OS, arch))
		case "oc":
			extractPath, _ = filepath.Abs(file + "/oc")
		default:
			// Generic extraction - look for binary with tool name
			extractPath, _ = filepath.Abs(file + "/" + toolConfig.Name)
		}

		if strings.Contains(osArch, "windows") {
			extractPath += exe
		}

		body, err = ioutil.ReadFile(extractPath)
		if err != nil {
			return err
		}
	} else if toolConfig.CompressedType == "zip" {
		// Handle zip files (future extension)
		// For now, treat as binary
	}
	// For "binary" type, body is already the binary content

	err = ioutil.WriteFile(fileName, body, 0750)
	if err != nil {
		return err
	}

	return nil
}

// =============================================================================
// VERSION MANAGEMENT
// =============================================================================

func sortVersions(versions []*version.Version, allReleases bool, allVersions bool) ([]*version.Version, error) {
	var numberOfVersion int
	var finalVersions []*version.Version

	sort.Sort(sort.Reverse(version.Collection(versions)))

	if allVersions {
		numberOfVersion = len(versions) - 1
	} else {
		numberOfVersion = 19
	}

	for i := 0; i <= numberOfVersion; i++ {
		if i == len(versions) {
			break
		}

		if allReleases {
			finalVersions = append(finalVersions, versions[i])
		} else {
			if !strings.ContainsAny(versions[i].String(), "beta") &&
				!strings.ContainsAny(versions[i].String(), "alpha") &&
				!strings.ContainsAny(versions[i].String(), "rc") {
				finalVersions = append(finalVersions, versions[i])
			} else {
				versions = append(versions[:i], versions[i+1:]...)
				i--
			}
		}
	}

	return finalVersions, nil
}

func printVersions(versions []*version.Version) {
	for _, element := range versions {
		fmt.Println(element)
	}
}

func getLocalVersions(binary string) ([]*version.Version, error) {
	var versions []*version.Version

	home, _ := homedir.Dir()
	binDir, _ := filepath.Abs(fmt.Sprintf("%s/.bin/%s-v*", home, binary))
	matches, _ := filepath.Glob(binDir)

	for _, match := range matches {
		v := strings.Split(match, string(os.PathSeparator))
		vs := strings.Replace(v[len(v)-1], binary+"-v", "", 1)

		if runtime.GOOS == "windows" {
			vs = strings.Replace(vs, ".exe", "", 1)
		}

		ver, err := version.NewVersion(vs)
		if err != nil {
			return versions, err
		}

		versions = append(versions, ver)
	}

	return versions, nil
}

func getRemoteVersions(endpoint string) ([]*version.Version, error) {
	var versions []*version.Version
	var defaultHTTPTimeout time.Duration = time.Second * 10
	var client = http.Client{Timeout: defaultHTTPTimeout}

	resp, err := client.Get(endpoint + "1")
	if err != nil {
		return versions, err
	}

	defer resp.Body.Close()

	forbidden := 403
	if resp.StatusCode == forbidden {
		fmt.Println("The request to Github's API failed, sorry.")
		fmt.Println("You may still install the version you want, if you know it. It will always go as X.Y.Z.")
		fmt.Println("The complete request response is ", resp)
		os.Exit(1)
	}

	lastPage, err := getLastPage(resp.Header.Get("Link"))
	if err != nil {
		return versions, err
	}

	for page := 2; page <= lastPage; page++ {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return versions, err
		}

		rel := []Page{}

		err = json.Unmarshal(body, &rel)
		if err != nil {
			return versions, err
		}

		if page != lastPage {
			resp, err = client.Get(endpoint + strconv.Itoa(page))
			if err != nil {
				return versions, err
			}
			defer resp.Body.Close()
		}

		for _, element := range rel {
			v, err := version.NewVersion(element.Release)
			if err != nil {
				return versions, err
			}
			versions = append(versions, v)
		}
	}

	return versions, nil
}

// =============================================================================
// WRAPPER FUNCTIONALITY
// =============================================================================

func runWrapper(toolName string) {
	// Get tool configuration
	toolConfig, exists := getTool(toolName)
	if !exists {
		fmt.Printf("Unsupported tool: %s\n", toolName)
		fmt.Printf("Supported tools: %v\n", getAllTools())
		os.Exit(1)
	}

	home, _ := homedir.Dir()
	var binPath string = fmt.Sprintf("%s/.bin", home)
	var defaultVersion string = fmt.Sprintf("%s/.%s-version", binPath, toolName)
	var localVersion string = fmt.Sprintf(".%s_version", toolName)
	var rawVersion []byte
	var finalVersion string
	var fileExt string
	var err error

	defaultVersion, _ = filepath.Abs(defaultVersion)
	localVersion, _ = filepath.Abs(localVersion)

	if _, err := os.Stat(localVersion); err == nil {
		rawVersion, err = ioutil.ReadFile(localVersion)
		if err != nil {
			fmt.Println("File reading error", err)
			return
		}
	} else {
		if _, err := os.Stat(defaultVersion); err != nil {
			d := []byte("auto\n")
			err = ioutil.WriteFile(defaultVersion, d, 0750)
			if err != nil {
				return
			}
		}

		rawVersion, err = ioutil.ReadFile(defaultVersion)
		if err != nil {
			fmt.Println("File reading error", err)
			return
		}
	}

	finalVersion = strings.Trim(string(rawVersion), "\n")

	if runtime.GOOS == "windows" {
		fileExt = ".exe"
	}

	// Handle auto-detection based on tool configuration
	if finalVersion == "auto" && toolConfig.SupportsAuto {
		version, err := kubeGetVersion()
		if err != nil {
			fmt.Println("Error getting kubernetes version: ", err)
			return
		}

		bin := fmt.Sprintf("%s/%s-v%s%s", binPath, toolName, version, fileExt)
		bin, _ = filepath.Abs(bin)

		if !fileExists(bin) {
			args := []string{toolName, "install", version}
			// Use the current binary for installation
			cmd := exec.Command("binarius"+fileExt, args...)
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			checkGenericError(err)
		}

		finalVersion = version
	}

	bin := fmt.Sprintf("%s/%s-v%s", binPath, toolName, finalVersion)
	bin, _ = filepath.Abs(bin)
	bin += fileExt

	cmd := exec.Command(bin, os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// =============================================================================
// COMMAND IMPLEMENTATIONS
// =============================================================================

func installCommand(toolConfig ToolConfig, version string) {
	var err error
	var osArch string

	// Check if os/arch is supported
	osArch, err = getOSArch()
	if err, ok := err.(*OSArchError); ok {
		if err.Err == "os not supported" {
			fmt.Printf("The OS '%s' is not supported.\n", err.OS)
		}
		if err.Err == "arch not supported" {
			fmt.Printf("The arch '%s' is not supported.\n", err.Arch)
		}
		os.Exit(0)
	}

	// Set base bin directory
	home, _ := homedir.Dir()
	fileName := fmt.Sprintf("%s/.bin/%s-v%s", home, toolConfig.Name, version)
	fileName, _ = filepath.Abs(fileName)

	if strings.Contains(osArch, "windows") {
		fileName += exe
	}

	// Check if binary exists locally
	if fileExists(fileName) {
		fmt.Printf("The version %s is already installed!\n", version)
		os.Exit(0)
	}

	// Download binary
	body, err := downloadBinaryWithConfig(version, toolConfig)
	// Check for errors when downloading the binary
	if err, ok := err.(*DownloadBinaryError); ok {
		if err.Err == "binary not found" {
			fmt.Println("The binary was not found. The url is:")
			fmt.Println(err.URL)
			os.Exit(0)
		}
		if err.Err == "unhandled error" {
			fmt.Println("There was an unhandled error downloading the binary, sorry:")
			fmt.Printf("Url: %s\n", err.URL)
			fmt.Printf("Error: %s\n", err.Body)
		}
	}

	checkGenericError(err)

	err = saveBinaryWithConfig(fileName, body, toolConfig)
	checkGenericError(err)
	fmt.Printf("Done! Saving it at %s.\n", fileName)
}

func listLocalCommand(toolConfig ToolConfig, allReleases, allVersions bool) {
	versions, err := getLocalVersions(toolConfig.Name)
	checkGenericError(err)

	versions, err = sortVersions(versions, allReleases, allVersions)
	checkGenericError(err)

	printVersions(versions)
}

func listRemoteCommand(toolConfig ToolConfig, allReleases, allVersions bool) {
	versions, err := getRemoteVersions(toolConfig.VersionsAPI)
	checkGenericError(err)

	versions, err = sortVersions(versions, allReleases, allVersions)
	checkGenericError(err)

	printVersions(versions)
}

func uninstallCommand(toolConfig ToolConfig, version string) {
	// Set base bin directory
	home, _ := homedir.Dir()
	fileName := fmt.Sprintf("%s/.bin/%s-v%s", home, toolConfig.Name, version)
	fileName, _ = filepath.Abs(fileName)

	if runtime.GOOS == "windows" {
		fileName += exe
	}

	// Check if binary exists locally
	if fileExists(fileName) {
		err := os.Remove(fileName)
		checkGenericError(err)
		fmt.Printf("Done! %s version uninstalled from %s.\n", version, fileName)
		os.Exit(0)
	}

	fmt.Printf("The version %s was already uninstalled! Doing nothing.\n", version)
}

func useCommand(toolConfig ToolConfig, version string) {
	home, _ := homedir.Dir()
	binPath := fmt.Sprintf("%s/.bin", home)
	defaultBin := fmt.Sprintf("%s/.%s-version", binPath, toolConfig.Name)
	defaultBin, _ = filepath.Abs(defaultBin)

	err := ioutil.WriteFile(defaultBin, []byte(version), 0750)
	checkGenericError(err)

	fmt.Printf("Done! Using %s version.\n", version)
}

// =============================================================================
// COBRA COMMANDS
// =============================================================================

func createToolCommand(toolName string, toolConfig ToolConfig) *cobra.Command {
	toolCmd := &cobra.Command{
		Use:   toolName,
		Short: fmt.Sprintf("Manage %s versions", toolConfig.Description),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
		},
	}

	// Install command
	installCmd := &cobra.Command{
		Use:   "install [version]",
		Short: "Install binary",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			installCommand(toolConfig, args[0])
		},
	}

	// List command with subcommands
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lists local and remote versions",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
			}
		},
	}
	listCmd.PersistentFlags().Bool("all-releases", false, "return all releases, including alpha, beta and rc releases")
	listCmd.PersistentFlags().Bool("all-versions", false, "return all versions")

	localCmd := &cobra.Command{
		Use:   "local",
		Short: "List local versions",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			allReleases, _ := cmd.Flags().GetBool("all-releases")
			allVersions, _ := cmd.Flags().GetBool("all-versions")
			listLocalCommand(toolConfig, allReleases, allVersions)
		},
	}

	remoteCmd := &cobra.Command{
		Use:   "remote",
		Short: "List remote versions",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			allReleases, _ := cmd.Flags().GetBool("all-releases")
			allVersions, _ := cmd.Flags().GetBool("all-versions")
			listRemoteCommand(toolConfig, allReleases, allVersions)
		},
	}

	listCmd.AddCommand(localCmd)
	listCmd.AddCommand(remoteCmd)

	// Uninstall command
	uninstallCmd := &cobra.Command{
		Use:   "uninstall [version]",
		Short: "Uninstall binary",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			uninstallCommand(toolConfig, args[0])
		},
	}

	// Use command
	useCmd := &cobra.Command{
		Use:   "use [version]",
		Short: "Set the default version to use",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			useCommand(toolConfig, args[0])
		},
	}

	// Tool-specific version command
	toolVersionCmd := &cobra.Command{
		Use:   "version",
		Short: "Outputs the current build information",
		Run: func(cmd *cobra.Command, args []string) {
			shortened := false
			output := "yaml"
			resp := goVersion.FuncWithOutput(shortened, Version, Commit, Date, output)
			fmt.Print(resp)
		},
	}

	toolCmd.AddCommand(installCmd)
	toolCmd.AddCommand(listCmd)
	toolCmd.AddCommand(uninstallCmd)
	toolCmd.AddCommand(useCmd)
	toolCmd.AddCommand(toolVersionCmd)

	return toolCmd
}

func createRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "binarius",
		Short: "Universal binary version manager for command-line tools",
		Long:  "Manage versions of command-line tools from a single binary",
	}

	// Add tool subcommands
	for toolName, toolConfig := range supportedTools {
		rootCmd.AddCommand(createToolCommand(toolName, toolConfig))
	}

	// Add global commands
	listAllCmd := &cobra.Command{
		Use:   "list [tool]",
		Short: "List versions for all tools or specific tool",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				// List all tools
				fmt.Println("Supported tools:")
				for _, toolName := range getAllTools() {
					toolConfig, _ := getTool(toolName)
					fmt.Printf("  %s - %s\n", toolName, toolConfig.Description)
				}
			} else {
				// List specific tool versions
				toolName := args[0]
				if err := validateTool(toolName); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				toolConfig, _ := getTool(toolName)
				listLocalCommand(toolConfig, false, false)
			}
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show binarius version",
		Run: func(cmd *cobra.Command, args []string) {
			shortened := false
			output := "yaml"
			if len(args) > 0 {
				if args[0] == "--short" {
					shortened = true
				}
				if args[0] == "--json" || (len(args) > 1 && args[1] == "--json") {
					output = "json"
				}
			}
			resp := goVersion.FuncWithOutput(shortened, Version, Commit, Date, output)
			fmt.Print(resp)
		},
	}
	versionCmd.Flags().Bool("short", false, "Print just the version number.")
	versionCmd.Flags().String("output", "yaml", "Output format. One of 'yaml' or 'json'.")

	rootCmd.AddCommand(listAllCmd)
	rootCmd.AddCommand(versionCmd)

	return rootCmd
}

// =============================================================================
// MODE DETECTION AND MAIN LOGIC
// =============================================================================

func detectExecutionMode() (mode string, toolName string) {
	execName := filepath.Base(os.Args[0])

	// Check if executed as tool symlink
	if _, exists := getTool(execName); exists {
		return "wrapper", execName
	}

	// Check if executed as manager
	if execName == "binarius" || strings.HasPrefix(execName, "binarius") {
		return "manager", ""
	}

	// Check if executed as binarius-wrapper with tool argument
	if strings.Contains(execName, "wrapper") && len(os.Args) >= 2 {
		if _, exists := getTool(os.Args[1]); exists {
			return "wrapper", os.Args[1]
		}
	}

	return "unknown", ""
}

func runManagerMode() {
	// Ensure .bin directory exists
	home, _ := homedir.Dir()
	_ = os.MkdirAll(home+"/.bin", os.ModePerm)

	rootCmd := createRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Println("Usage:")
	fmt.Println("  Manager mode: binarius [tool] [command] [args...]")
	fmt.Println("  Wrapper mode: Create symlinks kubectl, helm, oc -> binarius")
	fmt.Printf("Supported tools: %v\n", getAllTools())
}

func main() {
	mode, toolName := detectExecutionMode()

	switch mode {
	case "wrapper":
		// Remove tool name from args for wrapper execution if called as binarius-wrapper
		if strings.Contains(filepath.Base(os.Args[0]), "wrapper") {
			os.Args = append(os.Args[:1], os.Args[2:]...)
		}
		runWrapper(toolName)
	case "manager":
		runManagerMode()
	default:
		showUsage()
		os.Exit(1)
	}
}