use std::path::PathBuf;

use clap::Parser;
use prometheus::{Encoder, TextEncoder};
use solidity_ibc_eureka_relayer::{
    cli::{
        cmd::{Commands, RelayerCli},
        config::RelayerConfig,
    },
    core::builder::RelayerBuilder,
    modules::{
        cosmos_to_cosmos::CosmosToCosmosRelayerModule, cosmos_to_eth::CosmosToEthRelayerModule,
        eth_to_cosmos::EthToCosmosRelayerModule,
    },
};
use warp::Filter;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let cli = RelayerCli::parse();
    match cli.command {
        Commands::Start(args) => {
            let config_path = PathBuf::from(args.config);
            let config_bz = std::fs::read(config_path)?;
            let config: RelayerConfig = serde_json::from_slice(&config_bz)?;

            // Initialize the logger with log level.
            tracing_subscriber::fmt::fmt()
                .with_max_level(config.server.log_level())
                .init();

            // Build the relayer server.
            let mut relayer_builder = RelayerBuilder::default();
            relayer_builder.add_module(CosmosToEthRelayerModule);
            relayer_builder.add_module(CosmosToCosmosRelayerModule);
            relayer_builder.add_module(EthToCosmosRelayerModule);

            // Start the metrics server.
            tokio::spawn(async {
                let metrics_route = warp::path("metrics").map(|| {
                    let encoder = TextEncoder::new();
                    let metric_families = prometheus::gather();
                    let mut buffer = Vec::new();
                    encoder.encode(&metric_families, &mut buffer).unwrap();
                    String::from_utf8(buffer).unwrap()
                });

                tracing::info!("Metrics available at http://0.0.0.0:9000/metrics");
                warp::serve(metrics_route).run(([0, 0, 0, 0], 9000)).await;
            });

            // Start the relayer server.
            relayer_builder.start(config).await?;

            Ok(())
        }
    }
}
