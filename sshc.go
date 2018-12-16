package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kevinburke/ssh_config"
	"github.com/spf13/cobra"
)

// Variables holding data for Cobra flags used in the command-line interface
var (
	UserConfigFile      string
	CopyDestinationPath string
	JSONOutput          bool
	UseANSIColor        bool
	RegexIgnoreCase     bool
)

func FindSpecificHost(c *ssh_config.Config, alias string) *ssh_config.Host {
	for _, host := range NonWildcardHosts(c) {
		if host.Matches(alias) {
			return host
		}
	}
	return nil
}

func NonWildcardHosts(c *ssh_config.Config) []*ssh_config.Host {
	var filtered []*ssh_config.Host
	for _, host := range c.Hosts {
		if !strings.ContainsRune(host.Patterns[0].String(), '*') {
			filtered = append(filtered, host)
		}
	}
	return filtered
}

func LocateSSHConfig() (string, error) {
	localconf := UserConfigFile
	if strings.HasPrefix(UserConfigFile, "~") {
		// Let's try splitting the ~ off the path. This is an ascii character so we won't need to find the size of the rune
		split := UserConfigFile[1:]
		localconf = filepath.Join(os.Getenv("HOME"), split)
	}
	// At this point, UserConfigFile should be an expanded path to the file
	_, err := os.Stat(localconf)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ColorError(fmt.Sprintf("SSH config not found at path %v", localconf), Red)
		}
		return "", err
	}
	return localconf, nil
}

func ParseConfig() (*ssh_config.Config, error) {
	config_location, err := LocateSSHConfig()
	if err != nil {
		return nil, err
	}
	fh, err := os.Open(config_location)
	if err != nil {
		return nil, err
	}
	cfg, err := ssh_config.Decode(fh)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

/*
 *
 * Cobra Stuff
 *
 */

// find subcommand

func RunFind(needle string) error {
	var startColor, endColor string
	results := make(map[string][]int)
	// Option parsing
	if UseANSIColor {
		startColor = GetEscape(Magenta)
		endColor = GetEscape(Reset)
	}
	if RegexIgnoreCase {
		needle = "(?i)" + needle
	}
	cfg, err := ParseConfig()
	if err != nil {
		return err
	}
	r, err := regexp.Compile(needle)
	if err != nil {
		return err
	}
	// Iterate through all Hosts in the config file
	for _, host := range cfg.Hosts {
		full_block := host.String()
		match := r.FindStringIndex(full_block)
		// Append to results as (Host String, indicies) where the indicies are start & end of the match
		if match != nil {
			results[full_block] = match
		}
	}
	// Iterate through the results, printing the full host blocks highlighted at the match location
	for host_block, indicies := range results {
		fmt.Print(host_block[:indicies[0]] + startColor + host_block[indicies[0]:indicies[1]] + endColor + host_block[indicies[1]:])
	}
	return nil
}

func NewFindCommand() *cobra.Command {
	find := &cobra.Command{
		Use:     "find",
		Aliases: []string{"grep", "search"},
		Short:   "Search through your configuration for a regex, return the whole Host block which matches it",
		Example: "  sshc find 'IdentityFile.*rsa'\n  sshc find '10\\.0\\.1\\..'",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("find requires one argument, a regex to search for in your ssh config\n")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := RunFind(args[0]); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	find.Flags().BoolVarP(&RegexIgnoreCase, "ignore-case", "i", false, "Make the regex case-insensitive")
	return find
}

// copy-def subcommand

func RunCopy(hostname string, remote_uri string) error {
	cfg, err := ParseConfig()
	if err != nil {
		return err
	}
	hostdef := FindSpecificHost(cfg, hostname)
	if hostdef == nil {
		return ColorError("Host argument doesn't match any definitions in your SSH config file", Red)
	}
	// Start constructing oneliner that will copy this definition
	// Need to create literal "\n" in the string so it'll be expanded properly on the other side
	shell_escaped_hostdef := strings.Replace(hostdef.String(), "\n", "\\n", -1)
	oneliner := `bash -c 'echo -e "` + shell_escaped_hostdef + `" >> ` + CopyDestinationPath + "'"
	process := exec.Command("ssh", remote_uri, oneliner)
	process.Stdin = os.Stdin
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr
	err = process.Start()
	if err != nil {
		return err
	}
	err = process.Wait()
	if err != nil {
		return err
	}
	return nil
}

func NewCopyCommand() *cobra.Command {
	copydef := &cobra.Command{
		Use:     "copy",
		Aliases: []string{"copy-def"},
		Short:   "Copy a definition from your local ssh config to the ssh config on a remote host",
		Long: `The copy/copy-def command takes a definition from your local SSH config and copies it into the SSH config of a remote host.
It is implemented by opening a bash shell on the remote host that appends the definition to the file in question. The file will never be truncated.
This command is vulnerable to shell injection if you craft a malicious remote-path argument, so be careful with this on an untrusted system.`,
		Example: "  sshc copy <Host from config> <user@remote_host:port>",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("copy requires two arguments, the Host to copy from your SSH config file and a SSH connection string for the remote host it should be copied to.\n")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := RunCopy(args[0], args[1]); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	copydef.Flags().StringVarP(&CopyDestinationPath, "remote-path", "p", "$HOME/.ssh/config", "Path of file to which the host definition should be appended")
	return copydef
}

// edit subcommand

func NewEditCommand() *cobra.Command {
	edit := &cobra.Command{
		Use:     "edit",
		Aliases: []string{"e"},
		Short:   "Open your config file in $EDITOR",
		Run: func(cmd *cobra.Command, args []string) {
			config_location, err := LocateSSHConfig()
			if err != nil {
				log.Fatalln(err)
			}
			editor, exists := os.LookupEnv("EDITOR")
			if !exists {
				editor = "vim"
			}
			process := exec.Command(editor, config_location)
			process.Stdin = os.Stdin
			process.Stdout = os.Stdout
			process.Stderr = os.Stderr
			err = process.Start()
			if err != nil {
				log.Fatalln(err)
			}
			err = process.Wait()
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	return edit
}

// hosts subcommand

func RunHosts() error {
	cfg, err := ParseConfig()
	if err != nil {
		return err
	}
	for _, host := range cfg.Hosts {
		for _, p := range host.Patterns {
			fmt.Println(p)
		}
	}
	return nil
}

func NewHostsCommand() *cobra.Command {
	hosts := &cobra.Command{
		Use:     "hosts",
		Aliases: []string{"remotes"},
		Short:   "Show all hosts defined in your configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if err := RunHosts(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	return hosts
}

// get subcommand

func RunGet(hostname string) error {
	cfg, err := ParseConfig()
	if err != nil {
		return err
	}
	hostdef := FindSpecificHost(cfg, hostname)
	if hostdef == nil {
		return ColorError("No matching hosts", Red)
	}
	if JSONOutput {
		output := map[string]interface{}{}
		for _, n := range hostdef.Nodes {
			trimmed := strings.TrimLeft(n.String(), " \t")
			if trimmed == "" {
				continue
			}
			spl := strings.SplitN(trimmed, " ", 2)
			output[spl[0]] = spl[1]
		}
		blob, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", blob)
		return nil
	}
	fmt.Println(hostdef)
	return nil
}

func NewGetCommand() *cobra.Command {
	get := &cobra.Command{
		Use:     "get",
		Short:   "Return a single host definition from your ssh config",
		Long:    `Find a single host definition within ~/.ssh/config and return it. The one positional argument is the name of the Host pattern to match.`,
		Example: "  sshc get <Host from config>",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("get requires one argument, a Host to return from your SSH config file\n")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := RunGet(args[0])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	get.Flags().BoolVarP(&JSONOutput, "json-output", "j", false, "Output this host in JSON format")
	return get
}

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "sshc",
		Short: "Your ~/.ssh/config swiss-army knife",
		Long:  `This tool understands the format of your SSH configuration and lets you pull out information from it and copy definitions to remote hosts`,
	}
	root.PersistentFlags().StringVarP(&UserConfigFile, "config", "c", "~/.ssh/config", "Path to ssh config file")
	root.PersistentFlags().BoolVar(&UseANSIColor, "color", false, "Use ANSI terminal colors")
	return root
}

func main() {
	cmd := NewRootCmd()
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewHostsCommand())
	cmd.AddCommand(NewEditCommand())
	cmd.AddCommand(NewCopyCommand())
	cmd.AddCommand(NewFindCommand())
	cmd.Execute()
}
