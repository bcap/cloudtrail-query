package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cloudtrailquery "github.com/bcap/cloudtrailquery"
)

type args struct {
	Query           string `arg:"positional,required" help:"CloudTrail SQL query. Use - to read query from stdin"`
	Profile         string `arg:"-p,--profile" help:"Which AWS profile to use" default:"default"`
	NoNormalization bool   `arg:"-N,--no-normalization" help:"Do not normalize json results. Normalization consists of fixing value types, like replacing empty strings into null values, \"true\" strings into true values, etc"`
	NoExpansion     bool   `arg:"-E,--no-expansion" help:"Do not try to transform json-like values found in CloudTrail event columns into json objects. When expansion is turned on (the default), parsed column values have the \"__parsed\" suffix"`
}

func main() {
	args := args{}
	arg.MustParse(&args)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(args.Profile))
	failOnErr(err)

	query := args.Query
	if query == "-" {
		data, err := io.ReadAll(os.Stdin)
		failOnErr(err)
		query = string(data)
	}

	err = cloudtrailquery.QueryStream(
		ctx,
		cloudtrail.NewFromConfig(cfg),
		query,
		os.Stdout,
		cloudtrailquery.Expand(!args.NoExpansion),
		cloudtrailquery.Normalize(!args.NoNormalization),
	)
	failOnErr(err)
}

func failOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
