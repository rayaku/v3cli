package commands

import (
	"fmt"
	"github.com/nuclio/logger"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/v3io/v3cli/pkg/config"
	"github.com/v3io/v3cli/pkg/utils"
	"github.com/v3io/v3io-go-http"
	"io"
	"os"
	"strings"
)

type RootCommandeer struct {
	logger      logger.Logger
	cmd         *cobra.Command
	v3ioPath    string
	dirPath     string
	v3iocfg     *config.V3ioConfig
	cfgFilePath string
	logLevel    string
	container   string
	username    string
	password    string
	out         io.Writer
}

func NewRootCommandeer() *RootCommandeer {
	commandeer := &RootCommandeer{}

	cmd := &cobra.Command{
		Use:     "v3cli [command] [container] [path] [flags]",
		Short:   "v3io command line utility",
		Long:    GetLongHelp("root"),
		Example: GetExample("root"),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			file, err := cmd.Flags().GetString("input-file")
			if err == nil {
				err = GetInFile(file)
				if err != nil {
					return err
				}
			}
			if file == "" {
				InFile = os.Stdin
			} else {
				InFile, err = os.Open(file)
				if err != nil {
					return fmt.Errorf("Failed to open input file: %s\n", err)
				}
			}
			commandeer.out = InFile
			if len(args) > 0 {
				Container = args[0]
				commandeer.container = args[0]
			}
			if len(args) > 1 {
				commandeer.dirPath = args[1]
				Path = args[1]
			}
			if (cmd.Name() != "ls" && cmd.Name() != "complete" && cmd.Name() != "bash") && len(args) < 1 {
				return fmt.Errorf("Please specify container Name/Id, Path and parameters !\n")
			}

			return nil
		},
		SilenceUsage: true,
	}

	defaultV3ioServer := os.Getenv("V3IO_SERVICE_URL")

	cmd.PersistentFlags().StringVarP(&commandeer.logLevel, "log-level", "v", "",
		"Verbose output. You can provide one of the following logging\nlevels as an argument for this flag by using the assignment\noperator ('='): \"debug\" | \"info\" | \"warn\" | \"error\".\nFor example: -v=info. The default log level when using this\nflag without an argument is \""+config.DefaultLoggingLevel+"\".")
	cmd.PersistentFlags().Lookup("log-level").NoOptDefVal = config.DefaultLoggingLevel

	cmd.PersistentFlags().StringVarP(&commandeer.v3ioPath, "server", "s", defaultV3ioServer,
		"Web-gateway (web-APIs) service endpoint of an instance of\nthe Iguazio Continuous Data Platfrom, of the format\n\"<IP address>:<port number=8081>\". Examples: \"localhost:8081\"\n(when running on the target platform); \"192.168.1.100:8081\".")
	cmd.PersistentFlags().StringVarP(&commandeer.cfgFilePath, "config", "g", "",
		"Path to a YAML TSDB configuration file. When this flag isn't\nset, the CLI checks for a "+config.DefaultConfigurationFileName+" configuration file in the\ncurrent directory. CLI flags override file confiugrations.\nExample: \"~/cfg/my_v3io_tsdb_cfg.yaml\".")
	cmd.PersistentFlags().StringVarP(&commandeer.container, "container", "c", "",
		"The name of an Iguazio Continous Data Platform data container\nin which to create the TSDB table. Example: \"bigdata\".")
	cmd.PersistentFlags().StringVarP(&commandeer.username, "username", "u", "",
		"Username of an Iguazio Continous Data Platform user.")
	cmd.PersistentFlags().StringVarP(&commandeer.password, "password", "p", "",
		"Password of the configured user (see -u|--username).")

	// Add children
	cmd.AddCommand(
		NewCmdLS(commandeer).cmd,
		NewCmdGet(),
		NewCmdPut(),
		NewCmdPutitem(),
		NewCmdGetitem(),
		NewCmdGetitems(commandeer).cmd,
		NewCmdGetrecord(),
		NewCmdPutrecord(),
		NewCmdCreatestream(),
		NewCmdComplete(),
		NewCmdBash(),
		NewCmdIngest(),
	)

	commandeer.cmd = cmd

	return commandeer
}

// Execute the command using os.Args
func (rc *RootCommandeer) Execute() error {
	return rc.cmd.Execute()
}

// Return the underlying Cobra command
func (rc *RootCommandeer) GetCmd() *cobra.Command {
	return rc.cmd
}

// Generate Markdown files in the target path
func (rc *RootCommandeer) CreateMarkdown(path string) error {
	return doc.GenMarkdownTree(rc.cmd, path)
}

func (rc *RootCommandeer) initialize() error {
	cfg, err := config.GetOrLoadFromFile(rc.cfgFilePath)
	if err != nil {
		// Display an error if we fail to load a configuration file
		if rc.cfgFilePath == "" {
			return errors.Wrap(err, "Failed to load the TSDB configuration.")
		} else {
			return errors.Wrap(err, fmt.Sprintf("Failed to load the TSDB configuration from '%s'.", rc.cfgFilePath))
		}
	}
	return rc.populateConfig(cfg)
}

func (rc *RootCommandeer) populateConfig(cfg *config.V3ioConfig) error {

	if rc.username != "" {
		cfg.Username = rc.username
	}

	if rc.password != "" {
		cfg.Password = rc.password
	}

	if rc.v3ioPath != "" {
		// Check for username and password in the web-gateway service endpoint
		// (supported for backwards compatibility)
		if i := strings.LastIndex(rc.v3ioPath, "@"); i > 0 {
			usernameAndPassword := rc.v3ioPath[0:i]
			rc.v3ioPath = rc.v3ioPath[i+1:]
			if userpass := strings.Split(usernameAndPassword, ":"); len(userpass) > 1 {
				fmt.Printf("Debug: up0=%s up1=%s u=%s p=%s\n", userpass[0], userpass[1], rc.username, rc.password)
				if userpass[0] != "" && rc.username != "" {
					return fmt.Errorf("Username should only be defined once.")
				} else {
					cfg.Username = userpass[0]
				}

				if userpass[1] != "" && rc.password != "" {
					return fmt.Errorf("Password should only be defined once.")
				} else {
					cfg.Password = userpass[1]
				}
			} else {
				if usernameAndPassword != "" && rc.username != "" {
					return fmt.Errorf("Username should only be defined once.")
				} else {
					cfg.Username = usernameAndPassword
				}
			}
		}

		// Check for a container name in the in the web-gateway service endpoint
		// (supported for backwards compatibility)
		slash := strings.LastIndex(rc.v3ioPath, "/")
		if slash == -1 || len(rc.v3ioPath) <= slash+1 {
			if rc.container != "" {
				cfg.Container = rc.container
			} else if cfg.Container == "" {
				return fmt.Errorf("Missing the name of the TSDB's parent data container.")
			}
			cfg.WebApiEndpoint = rc.v3ioPath
		} else {
			cfg.WebApiEndpoint = rc.v3ioPath[0:slash]
			cfg.Container = rc.v3ioPath[slash+1:]
		}
	}
	if rc.container != "" {
		cfg.Container = rc.container
	}
	if cfg.WebApiEndpoint == "" || cfg.Container == "" {
		return fmt.Errorf("Not all required configuration information was provided. The endpoint of the web-gateway service, related username and password authentication credentials, the name of the TSDB parent data container, and the path to the TSDB table within the container, must be defined as part of the CLI command or in a configuration file.")
	}
	if rc.logLevel != "" {
		cfg.LogLevel = rc.logLevel
	}

	rc.v3iocfg = cfg
	return nil
}

func (rc *RootCommandeer) initV3io() (*v3io.Container, error) {

	if rc.container == "" {
		rc.container = rc.v3iocfg.Container
	}

	rc.logger, _ = utils.NewLogger(rc.v3iocfg.LogLevel)

	newContainer, err := utils.CreateContainer(
		rc.logger, rc.v3iocfg.WebApiEndpoint, rc.container, rc.v3iocfg.Username, rc.v3iocfg.Password, rc.v3iocfg.Workers)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize a data container.")
	}

	return newContainer, nil
}