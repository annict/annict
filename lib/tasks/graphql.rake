# frozen_string_literal: true

require "graphql/rake_task"

namespace :graphql do
  task dump_schema: :environment do
    GraphQL::RakeTask.new(schema_name: "AnnictSchema")
  end

  task build_docs: :environment do
    config = {
      filename: "#{Rails.root}/app/graphql/schema.graphql",
      output_dir: "../developers/graphql-api/reference/",
      base_url: "/graphql-api/reference"
    }
    GraphQLDocs.build(config)
  end
end
