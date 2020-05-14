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
            itemable_type: node["itemableType"].downcase,
            single: node["single"],
            activities_count: node["activitiesCount"],
            created_at: node["createdAt"],
            user: build_user(node["user"]),
            itemables: build_itemables(node.dig("activities", "nodes"))
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

    def build_itemables(activities)
      activities.map do |activity|
        itemable = activity["itemable"]

        case activity["itemableType"]
        when "EPISODE_RECORD"
          EpisodeRecordEntity.new(
            id: itemable["annictId"],
            rating_state: itemable["ratingState"]&.downcase,
            body_html: itemable["bodyHtml"],
            likes_count: itemable["likesCount"],
            comments_count: itemable["commentsCount"],
            work: build_work(itemable["work"]),
            episode: build_episode(itemable["episode"])
          )
        when "STATUS"
          StatusEntity.new(
            id: itemable["annictId"],
            kind: itemable["kind"].downcase,
            likes_count: itemable["likesCount"],
            work: build_work(itemable["work"])
          )
        when "WORK_RECORD"
          WorkRecordEntity.new(
            id: itemable["annictId"],
            rating_overall_state: itemable["ratingOverallState"]&.downcase,
            body_html: itemable["bodyHtml"],
            likes_count: itemable["likesCount"],
            work: build_work(itemable["work"]),
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
