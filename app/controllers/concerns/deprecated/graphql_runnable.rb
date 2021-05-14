# frozen_string_literal: true

module Deprecated::GraphqlRunnable
  def graphql_client(viewer: nil)
    @graphql_client ||= Annict::Deprecated::Graphql::InternalClient.new(
      viewer: viewer
    )
  end
end
