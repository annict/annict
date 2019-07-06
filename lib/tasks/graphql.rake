# frozen_string_literal: true

require "graphql/rake_task"

namespace :graphql do
  task :dump_schema, %i(version) => :environment do |_, args|
    version = args[:version]

    schema_definition = if version
      "#{version.classify}::AnnictSchema".constantize.to_definition
    else
      AnnictSchema.to_definition
    end

    schema_path = if version
      "app/graphql/#{version}/schema.graphql"
    else
      "app/graphql/schema.graphql"
    end

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
