# frozen_string_literal: true

module UserHome
  class FollowingActivitiesRepository < ApplicationRepository
    def fetch(username:, cursor:)
      result = graphql_client.execute(query, variables: { username: username, cursor: cursor.presence || "" })
      data = result.to_h.dig("data", "user", "followingActivities")

      {
        pagination: {
          end_cursor: data.dig("pageInfo", "endCursor"),
          has_next_page: data.dig("pageInfo", "hasNextPage")
        },
        activities: data["nodes"].map do |node|
          ActivityEntity.new(
            id: node["annictId"],
            resource_type: node["resourceType"].downcase,
            resources_count: node["resourcesCount"],
            single: node["single"],
            created_at: node["createdAt"],
            user: build_user(node["user"]),
            resources: build_resources(node["resources"])
          )
        end
      }
    end

    private

    def query
      load_query "user_home/following_activities.graphql"
    end

    def build_user(user)
      UserEntity.new(
        username: user["username"],
        name: user["name"],
        avatar_url: user["avatarUrl"]
      )
    end

    def build_resources(resources)
      resources.map do |resource|
        case resource["__typename"]
        when "EpisodeRecord"
          EpisodeRecordEntity.new(
            type: "episode_record",
            rating_state: resource["ratingState"]&.downcase,
            body_html: resource["bodyHtml"],
            likes_count: resource["likesCount"],
            comments_count: resource["commentsCount"],
            work: build_work(resource["work"]),
            episode: build_episode(resource["episode"])
          )
        when "Status"
          StatusEntity.new(
            type: "status",
            kind: resource["kind"].downcase,
            likes_count: resource["likesCount"],
            work: build_work(resource["work"])
          )
        when "WorkRecord"
          WorkRecordEntity.new(
            type: "work_record",
            rating_overall_state: resource["ratingOverallState"]&.downcase,
            body_html: resource["bodyHtml"],
            likes_count: resource["likesCount"],
            work: build_work(resource["work"]),
          )
        end
      end
    end

    def build_work(work)
      WorkEntity.new(
        id: work["annictId"],
        title: work["title"],
        title_en: work["titleEn"],
        image_url_1x: work.dig("image", "internalUrl1x"),
        image_url_2x: work.dig("image", "internalUrl2x")
      )
    end

    def build_episode(episode)
      EpisodeEntity.new(
        id: episode["annictId"],
        number_text: episode["numberText"],
        title: episode["title"],
        title_en: episode["titleEn"]
      )
    end
  end
end
