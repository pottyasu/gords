package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"fmt"
	"os/exec"
	"strings"
)

type dbCluster struct {
	DBInstanceIdentifier string
	IsClusterWriter      bool
}

type dbInstance struct {
	DBInstanceIdentifier string
	DBInstanceStatus     string
	DBInstanceClass      string
	Engine               string
	EndpointAdress       string
	EndpointPort         int64
	MasterUserName       string
	EndpointType         string
	EngineVersion        string
	DBName               string
}

// global variables for Flags
var (
	profileName string
	regionName  string
	userName    string
	catFlag     bool
)

// init cmd
func init() {
	// Handling flags
	rootCmd.PersistentFlags().StringVarP(&profileName, "profile", "p", "default","Select profile")
	rootCmd.PersistentFlags().StringVarP(&regionName, "region", "r", "ap-northeast-1","Select region")
	rootCmd.PersistentFlags().StringVarP(&userName, "username", "u", "MasterUserName","Select user name")
	rootCmd.PersistentFlags().BoolVarP(&catFlag, "cat", "c", false, "Output to command line")
	// Handling Config
	initViperConfig()
}

var rootCmd = &cobra.Command{
	Use:   "gords",
	Short: "The cli tool for connecting to Amazon RDS DB Instance",
	Run: func(cmd *cobra.Command, args []string) {
		response := getEndpoints(regionName, profileName)
		i := showPromptSelecter(response)
		dbConneceter(response[i])
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Load config from gords_config.yml
func initViperConfig() {
	// set default values
	viper.SetDefault("mysqlClient", "mysql")
	viper.SetDefault("mariaClient", "mysql")
	viper.SetDefault("postgresClient", "psql")
	viper.SetDefault("mssqlClient", "mssql-cli")
	viper.SetDefault("oracleClient", "sqlplus64")

	// read from config file
	viper.SetConfigType("yaml")
	viper.SetConfigName(".gords_config")
	viper.AddConfigPath("$HOME/.gords")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
}

// run DescribeDBInstances() & DescribeDBClusters
func getEndpoints(regionName string, profileName string) []dbInstance {
	// init sess
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: profileName,
		Config:  aws.Config{Region: aws.String(regionName)},
	}))
	// DescribeDBInstances
	svc := rds.New(sess)
	res, err := svc.DescribeDBInstances(nil)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBInstanceNotFoundFault:
				fmt.Println(rds.ErrCodeDBInstanceNotFoundFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
	// DescribeDBClusters
	res2, err2 := svc.DescribeDBClusters(nil)
	if err2 != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBInstanceNotFoundFault:
				fmt.Println(rds.ErrCodeDBInstanceNotFoundFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
	outputs := responseParser(res, res2)
	return outputs
}

// parse DescribeDBInstances/DescribeDBClusters Output
func responseParser(res *rds.DescribeDBInstancesOutput, res2 *rds.DescribeDBClustersOutput) []dbInstance {

	parsedCluster := make([]dbCluster, 0)
	parsedResponse := make([]dbInstance, 0)
	var temp dbInstance
	var temp2 dbCluster
	var endpointType string

	for _, c := range res2.DBClusters {
		for _, cm := range c.DBClusterMembers {
			temp2 = dbCluster{
				DBInstanceIdentifier: *cm.DBInstanceIdentifier,
				IsClusterWriter:      *cm.IsClusterWriter,
			}
			parsedCluster = append(parsedCluster, temp2)
		}
	}

	for _, r := range res.DBInstances {
		// checking DB Instance role
		endpointType = "Instance"
		// if DB instances is creating status
		if aws.StringValue(r.DBInstanceStatus) == "creating" {
			temp = dbInstance{
				DBInstanceIdentifier: *r.DBInstanceIdentifier,
				DBInstanceStatus:     *r.DBInstanceStatus,
				DBInstanceClass:      *r.DBInstanceClass,
				Engine:               *r.Engine,
				EndpointAdress:       "Wait For Create...",
				EndpointPort:         0000,
				MasterUserName:       *r.MasterUsername,
				EndpointType:         "Wait For Create...",
				EngineVersion:        *r.EngineVersion,
				DBName:               aws.StringValue(r.DBName),
			}
			parsedResponse = append(parsedResponse, temp)
			continue
		}
		// check DB Engine
		switch *r.Engine {
		// if aurora, check IsClusterWriter.
		case "aurora", "aurora-mysql", "aurora-postgresql":
			for _, s := range parsedCluster {
				if *r.DBInstanceIdentifier == s.DBInstanceIdentifier {
					if s.IsClusterWriter == true {
						endpointType = "Writer"
					} else {
						endpointType = "Reader"
					}
				}
			}
		default:
			// if DB Instance is ReadReplica
			if aws.StringValue(r.ReadReplicaSourceDBInstanceIdentifier) != "" {
				endpointType = "Replica"
			}
			// if DB Instance has ReadReplica/Aurora ReadReplica
			if len(aws.StringValueSlice(r.ReadReplicaDBInstanceIdentifiers)) > 0 || len(aws.StringValueSlice(r.ReadReplicaDBClusterIdentifiers)) > 0 {
				endpointType = "Master"
			}
		}
		temp = dbInstance{
			DBInstanceIdentifier: *r.DBInstanceIdentifier,
			DBInstanceStatus:     *r.DBInstanceStatus,
			DBInstanceClass:      *r.DBInstanceClass,
			Engine:               *r.Engine,
			EndpointAdress:       *r.Endpoint.Address,
			EndpointPort:         *r.Endpoint.Port,
			MasterUserName:       *r.MasterUsername,
			EndpointType:         endpointType,
			EngineVersion:        *r.EngineVersion,
			DBName:               aws.StringValue(r.DBName),
		}
		parsedResponse = append(parsedResponse, temp)
	}
	return parsedResponse
}

// show interactive prompt
func showPromptSelecter(outputs []dbInstance) int {

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F363  {{ .DBInstanceIdentifier | magenta }} ({{ .EndpointType| red }}) [{{ .DBInstanceStatus| green }}]",
		Inactive: "   {{ .DBInstanceIdentifier | cyan }} ({{ .EndpointType | red }}) [{{ .DBInstanceStatus| green }}]",
		Selected: "\U0001F363  {{ .DBInstanceIdentifier | red | cyan }} [{{ .DBInstanceStatus| green }}]",
		Details: `
--------- DB Details ----------
{{ "Engine / Version:" | faint }}	{{ .Engine }} / {{ .EngineVersion }}
{{ "Role:" | faint }}	{{ .EndpointType }}
{{ "DBInstanceStatus:" | faint }}	{{ .DBInstanceStatus }}
{{ "EndPoint:" | faint }}	{{ .EndpointAdress }}
{{ "MasterUserName / Port:" | faint }}	{{ .MasterUserName}} / {{ .EndpointPort }}
{{ "DBInstanceClass:" | faint }}	{{ .DBInstanceClass }}`,
	}
	searcher := func(input string, index int) bool {
		selecter := outputs[index]
		name := strings.Replace(strings.ToLower(selecter.DBInstanceIdentifier), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}
	prompt := promptui.Select{
		Label:     "Which One ",
		Items:     outputs,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}
	i, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
	}
	return i
}

// Connnecting DB Instance
func dbConneceter(selectedDBInstance dbInstance) {

	dbEngine := selectedDBInstance.Engine
	shCmd := ""

	if userName == "" {
		userName = selectedDBInstance.MasterUserName
	}
	// switch by engine type
	switch dbEngine {
	// MySQL , Aurora MySQL Compatible , MariaDB
	case "mysql", "aurora", "aurora-mysql":
		shCmd = fmt.Sprintf("%s -h %s -P %d -u %s -p",
			viper.GetString("mysqlClient"), selectedDBInstance.EndpointAdress, selectedDBInstance.EndpointPort, userName)
	// MariaDB
	case "mariadb":
		shCmd = fmt.Sprintf("%s -h %s -P %d -u %s -p",
			viper.GetString("mariaClient"), selectedDBInstance.EndpointAdress, selectedDBInstance.EndpointPort, userName)
	// PostgreSQL / Aurora PostgreSLQ Compatible
	case "postgres", "aurora-postgresql":
		shCmd = fmt.Sprintf("%s -h %s -p %d -U %s -d postgres",
			viper.GetString("postgresClient"), selectedDBInstance.EndpointAdress, selectedDBInstance.EndpointPort, userName)
	// Oracle
	case "oracle-ee", "oracle-se2", "oracle-se1", "oracle-se":
		shCmd = fmt.Sprintf("%s '%s@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=%s) (PORT=%d))(CONNECT_DATA=(SID=%s)))'",
			viper.GetString("oracleClient"), userName, selectedDBInstance.EndpointAdress, selectedDBInstance.EndpointPort, selectedDBInstance.DBName)
	case "sqlserver-ee", "sqlserver-se", "sqlserver-ex", "sqlserver-web":
		shCmd = fmt.Sprintf("%s -S tcp:%s,%d -U %s",
			viper.GetString("mssqlClient"), selectedDBInstance.EndpointAdress, selectedDBInstance.EndpointPort, userName)
	default:
		fmt.Println(" dbconnecter() Error : Unsupported Engine Type")
	}

	// check catFlag
	if catFlag == true {
		fmt.Println(shCmd)
		return
	}

	// execute
	fmt.Println("running... : ", shCmd)
	cmd := exec.Command("sh", "-c", shCmd)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}
