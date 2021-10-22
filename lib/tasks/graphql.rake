# frozen_string_literal: true

require "graphql/rake_task"

namespace :graphql do
  task :dump_schema, %i[version] => :environment do |_, args|
    version = args[:version]

    unless version
      puts "version required"
      next
    end

    schema_definition = "#{version.camelize}::AnnictSchema".constantize.to_definition
    schema_path = "app/graphql/#{version}/schema.graphql"

    File.write(Rails.root.join(schema_path), schema_definition)

    puts "Updated #{schema_path}"
  end

  task :build_reference, %i[version] => :environment do |_, args|
    version = args[:version]

    unless version
      puts "version required"
      next
    end

    GraphQLDocs.build(
      filename: "#{Rails.root}/app/graphql/#{version}/schema.graphql",
      output_dir: "./tmp/docs/graphql-api/#{version}/reference/",
      base_url: "/docs/graphql-api/#{version}/reference"
    )
  end

  task :copy_reference, %i[version] => :environment do |_, args|
    version = args[:version]

    unless version
      puts "version required"
      next
    end

    target_path = "../annict-developers/static/docs/graphql-api/#{version}/reference"
    system "mkdir -p #{target_path}"
    system "cp -rf ./tmp/docs/graphql-api/#{version}/reference/ #{target_path}"
  end
end
