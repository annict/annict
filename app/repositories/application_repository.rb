# frozen_string_literal: true

class ApplicationRepository
  def initialize(graphql_client:)
    @graphql_client = graphql_client
  end

  private

  attr_reader :graphql_client

  def load_query(path)
    File.read(Rails.root.join("app/graphql_queries/#{path}"))
  end
end
