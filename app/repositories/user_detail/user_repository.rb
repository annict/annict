# frozen_string_literal: true

module UserDetail
  class UserRepository < ApplicationRepository
    def fetch(username:)
      result = graphql_client.execute(query, variables: { username: username })
      node = result.to_h.dig("data", "user")

      UserEntity.new(
        username: node["username"],
        name: node["name"],
        description: node["description"],
        avatar_url: node["avatar_url"]
      )
    end

    private

    def query
      load_query "user_detail/fetch_user.graphql"
    end
  end
end
