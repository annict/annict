# frozen_string_literal: true

module UserDetail
  class UserRepository < ApplicationRepository
    def fetch(username:)
      result = graphql_client.execute(query, variables: { username: username })
      node = result.to_h.dig("data", "user")

      UserEntity.new(
        id: node["annictId"],
        username: node["username"],
        name: node["name"],
        description: node["description"],
        avatar_url: node["avatarUrl"],
        background_image_url: node["backgroundImageUrl"],
        display_supporter_badge: node["displaySupporterBadge"],
      )
    end

    private

    def query
      load_query "profile/fetch_user.graphql"
    end
  end
end
