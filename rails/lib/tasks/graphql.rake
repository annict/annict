# typed: false
# frozen_string_literal: true

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
end
