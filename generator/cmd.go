package generator

import "github.com/spf13/cobra"

func CreateCmd() *cobra.Command {
	genCmd := &cobra.Command{
		Use:   "gen",
		Short: "generator project structure",
		Long:  "generator project structure",
		Run: func(cmd *cobra.Command, _ []string) {
			modname, _ := cmd.Flags().GetString("modname")
			g := NewGen(modname, genOptions(cmd)...)
			g.Generate()
		},
	}
	setParmas(genCmd)
	return genCmd
}

func setParmas(cmd *cobra.Command) {
	cmd.Flags().String("modname", "", "project go module name, required")
	cmd.Flags().StringSlice("name", []string{}, "app name, multiple are given multiple times, required")
	cmd.Flags().String("mysql", "", "enable the MySQL component and specify the ORM, gorm, ent or sqlx")
	cmd.Flags().Bool("redis", false, "enable the redis component")
	cmd.Flags().Bool("cache", false, "enable the redis cache component")
	cmd.Flags().Bool("rocketmq_producer", false, "enable the rocket mq product component")
	cmd.Flags().Bool("rocketmq_consumer", false, "enable the rocket mq consumer component")
	cmd.Flags().Bool("trace", false, "enable the otel trace component")
	cmd.Flags().Bool("nacos_config", false, "enable the nacos configuration center component")
	cmd.Flags().Bool("nacos_naming", false, "enable the nacos service discovery and registration component")

	_ = cmd.MarkFlagRequired("modname")
	_ = cmd.MarkFlagRequired("name")
}

func genOptions(cmd *cobra.Command) []Option {
	var ops []Option
	appNam, _ := cmd.Flags().GetStringSlice("name")
	ops = append(ops, WithName(appNam...))

	mysql, _ := cmd.Flags().GetString("mysql")
	if mysql != "" {
		ops = append(ops, WithMysql(mysql))
	}
	redis, _ := cmd.Flags().GetBool("redis")
	if redis {
		ops = append(ops, WithRedis())
	}
	cache, _ := cmd.Flags().GetBool("cache")
	if cache {
		ops = append(ops, WithCache())
	}
	rocketmqProducer, _ := cmd.Flags().GetBool("rocketmq_producer")
	if rocketmqProducer {
		ops = append(ops, WithRocketMQProducer())
	}
	rocketmqConsumer, _ := cmd.Flags().GetBool("rocketmq_consumer")
	if rocketmqConsumer {
		ops = append(ops, WithRocketMQConsumer())
	}
	trace, _ := cmd.Flags().GetBool("trace")
	if trace {
		ops = append(ops, WithOtelTrace())
	}
	nacosConfig, _ := cmd.Flags().GetBool("nacos_config")
	if nacosConfig {
		ops = append(ops, WithNacosConfig())
	}
	nacosNaming, _ := cmd.Flags().GetBool("nacos_naming")
	if nacosNaming {
		ops = append(ops, WithNacosNaming())
	}

	return ops
}
