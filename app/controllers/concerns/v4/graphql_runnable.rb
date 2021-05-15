# frozen_string_literal: true

module V4::GraphqlRunnable
  def graphql_client(viewer: nil)
    @graphql_client ||= Annict::V4::Graphql::InternalClient.new(
      viewer: viewer
    )
  end
end
