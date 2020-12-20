# frozen_string_literal: true

class ApplicationRepository
  class ExecutionFailedError < StandardError; end

  module Types
    include Dry.Types()
  end

  class MutationError < Dry::Struct
    attribute :message, Types::String
  end

  def initialize(graphql_client:)
    @graphql_client = graphql_client
  end

  private

  attr_reader :graphql_client

  def file_name
    @file_name ||=
      "#{self.class.name.split('::').map { |str| str.camelize(:lower) }.join('/').delete_suffix('Repository')}.graphql"
  end

  def query_definition
    @query_definition ||= File.read(Rails.root.join("app", "lib", "annict", "graphql", "queries", file_name))
  end

  def mutation_definition
    @mutation_definition ||= File.read(Rails.root.join("app", "lib", "annict", "graphql", "mutations", file_name))
  end

  def camelized_variables(variables)
    variables.deep_transform_keys { |key| key.to_s.camelize(:lower) }
  end

  def query(variables: {})
    result = graphql_client.execute(query_definition, variables: camelized_variables(variables))
    validate! result
    result
  end

  def mutate(variables: {})
    result = graphql_client.execute(mutation_definition, variables: camelized_variables(variables))
    validate! result
    result
  end

  def validate!(result)
    if result["errors"]
      raise ExecutionFailedError, result.dig("errors", 0, "message")
    end
  end
end
