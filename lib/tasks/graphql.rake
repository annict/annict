# frozen_string_literal: true

require "graphql/rake_task"

namespace :graphql do
  task dump_schema: :environment do
    schema_definition = AnnictSchema.to_definition
    schema_path = "app/graphql/schema.graphql"
    File.write(Rails.root.join(schema_path), schema_definition)
    puts "Updated #{schema_path}"
  end

  task build_docs: :environment do
    config = {
      filename: "#{Rails.root}/app/graphql/schema.graphql",
      output_dir: "./tmp/docs/graphql-api/reference/",
      base_url: "/graphql-api/reference"
    }
    GraphQLDocs.build(config)
  end
end
