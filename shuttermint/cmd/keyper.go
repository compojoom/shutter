package cmd

import (
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/brainbot-com/shutter/shuttermint/keyper"
)

type KeyperConfig struct {
	ShuttermintURL string
	EthereumURL    string
	SigningKey     string
	ConfigContract string
}

// keyperCmd represents the keyper command
var keyperCmd = &cobra.Command{
	Use:   "keyper",
	Short: "Run a shutter keyper",
	Run: func(cmd *cobra.Command, args []string) {
		keyperMain()
	},
}

func init() {
	rootCmd.AddCommand(keyperCmd)
	keyperCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func readKeyperConfig() (KeyperConfig, error) {
	viper.SetEnvPrefix("KEYPER")
	viper.BindEnv("ShuttermintURL")
	viper.BindEnv("EthereumURL")
	viper.BindEnv("SigningKey")
	viper.SetDefault("ShuttermintURL", "http://localhost:26657")
	viper.SetDefault("EthereumURL", "ws://localhost:8545/websocket")
	defer func() {
		if viper.ConfigFileUsed() != "" {
			log.Printf("Read config from %s", viper.ConfigFileUsed())
		}
	}()
	var err error
	kc := KeyperConfig{}

	viper.AddConfigPath("$HOME/.config/shutter")
	viper.SetConfigName("keyper")
	viper.SetConfigType("toml")
	viper.SetConfigFile(cfgFile)

	err = viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found
		if cfgFile != "" {
			return kc, err
		}
	} else if err != nil {
		return kc, err // Config file was found but another error was produced
	}

	err = viper.Unmarshal(&kc)
	return kc, err
}

func keyperMain() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	kc, err := readKeyperConfig()
	if err != nil {
		log.Fatalf("Error reading the configuration file: %s\nPlease check your configuration.", err)
	}

	privateKey, err := crypto.HexToECDSA(kc.SigningKey)
	if err != nil {
		log.Fatalf("Error: bad signing key: %s\nPlease check your configuration.", err)
	}
	if !keyper.IsWebsocketURL(kc.EthereumURL) {
		log.Fatalf("Error: EthereumURL must start with ws:// or wss://\nPlease check your configuration.")
	}
	addr := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	log.Printf(
		"Starting keyper version %s with signing key %s, using %s for Shuttermint and %s for Ethereum",
		version,
		addr,
		kc.ShuttermintURL,
		kc.EthereumURL,
	)
	k := keyper.NewKeyper(privateKey, kc.ShuttermintURL, kc.EthereumURL)
	err = k.Run()
	if err != nil {
		panic(err)
	}
}
