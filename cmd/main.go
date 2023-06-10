package main

import (
	"context"
	"log"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cloudtrailquery "github.com/bcap/cloudtrail-query"
)

type args struct {
	Query           string `arg:"positional,required" help:"CloudTrail SQL query"`
	NoNormalization bool   `arg:"-N,--no-normalization" help:"Do not normalize json results"`
	NoExpansion     bool   `arg:"-E,--no-expansion" help:"Do not try to parse values into json objects"`
	Profile         string `arg:"-p,--profile" help:"Which aws profile to use" default:"default"`
}

func main() {
	args := args{}
	arg.MustParse(&args)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(args.Profile))
	failOnErr(err)

	err = cloudtrailquery.QueryStream(
		ctx,
		cloudtrail.NewFromConfig(cfg),
		args.Query,
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
