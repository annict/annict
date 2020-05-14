# frozen_string_literal: true

module UserHome
  class FollowingActivityGroupsRepository < ApplicationRepository
    def fetch(username:, cursor:)
      result = graphql_client.execute(query, variables: { username: username, cursor: cursor.presence || "" })
      data = result.to_h.dig("data", "user", "followingActivityGroups")

      {
        pagination: {
          end_cursor: data.dig("pageInfo", "endCursor"),
          has_next_page: data.dig("pageInfo", "hasNextPage")
        },
        activity_groups: data["nodes"].map do |node|
          ActivityGroupEntity.new(
            resource_type: node["resourceType"].downcase,
            single: node["single"],
            activities_count: node["activitiesCount"],
            created_at: node["createdAt"],
            user: build_user(node["user"]),
            resources: build_resources(node.dig("activities", "nodes"))
          )
        end
      }
    end

    private

    def query
      load_query "user_home/following_activity_groups.graphql"
    end

    def build_user(user)
      UserEntity.new(
        username: user["username"],
        name: user["name"],
        avatar_url: user["avatarUrl"]
      )
    end

    def build_resources(activities)
      activities.map do |activity|
        resource = activity["resource"]

        case activity["resourceType"]
        when "EPISODE_RECORD"
          EpisodeRecordEntity.new(
            id: resource["annictId"],
            rating_state: resource["ratingState"]&.downcase,
            body_html: resource["bodyHtml"],
            likes_count: resource["likesCount"],
            comments_count: resource["commentsCount"],
            work: build_work(resource["work"]),
            episode: build_episode(resource["episode"])
          )
        when "STATUS"
          StatusEntity.new(
            id: resource["annictId"],
            kind: resource["kind"].downcase,
            likes_count: resource["likesCount"],
            work: build_work(resource["work"])
          )
        when "WORK_RECORD"
          WorkRecordEntity.new(
            id: resource["annictId"],
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
