# frozen_string_literal: true

module UserHome
  module ActivityBuildable
    def build_page_info(page_info_node)
      PageInfoEntity.new(
        end_cursor: page_info_node["endCursor"],
        has_next_page: page_info_node["hasNextPage"]
      )
    end

    def build_activity_groups(activity_group_nodes)
      activity_group_nodes.map do |node|
        ActivityGroupEntity.new(
          id: node["id"],
          itemable_type: node["itemableType"].downcase,
          single: node["single"],
          activities_count: node["activitiesCount"],
          created_at: node["createdAt"],
          user: build_user(node["user"]),
          itemables: build_itemables(node.dig("activities", "nodes"), build_user(node["user"])),
          activities_page_info: PageInfoEntity.new(
            end_cursor: node.dig("activities", "pageInfo", "endCursor")
          )
        )
      end
    end

    def build_itemables(activities, user)
      activities.map do |activity|
        itemable = activity["itemable"]

        case activity["itemableType"]
        when "EPISODE_RECORD"
          build_episode_record(itemable, user)
        when "STATUS"
          build_status(itemable, user)
        when "WORK_RECORD"
          build_work_record(itemable, user)
        end
      end
    end

    def build_user(user)
      UserEntity.new(
        username: user["username"],
        name: user["name"],
        avatar_url: user["avatarUrl"]
      )
    end

    def build_episode_record(itemable, user)
      EpisodeRecordEntity.new(
        id: itemable["databaseId"],
        rating_state: itemable["ratingState"]&.downcase,
        body: itemable["body"],
        likes_count: itemable["likesCount"],
        comments_count: itemable["commentsCount"],
        work: build_work(itemable["work"]),
        episode: build_episode(itemable["episode"]),
        record: build_record(itemable["record"]),
        user: user
      )
    end

    def build_status(itemable, user)
      StatusEntity.new(
        id: itemable["databaseId"],
        kind: itemable["kind"].downcase,
        likes_count: itemable["likesCount"],
        work: build_work(itemable["work"]),
        user: user
      )
    end

    def build_work_record(itemable, user)
      WorkRecordEntity.new(
        id: itemable["databaseId"],
        rating_overall_state: itemable["ratingOverallState"]&.downcase,
        body: itemable["body"],
        likes_count: itemable["likesCount"],
        work: build_work(itemable["work"]),
        user: user
      )
    end

    def build_work(work)
      WorkEntity.new(
        id: work["databaseId"],
        title: work["title"],
        title_en: work["titleEn"],
        image_url_1x: work.dig("image", "internalUrl1x"),
        image_url_2x: work.dig("image", "internalUrl2x")
      )
    end

    def build_episode(episode)
      EpisodeEntity.new(
        id: episode["databaseId"],
        number_text: episode["numberText"],
        title: episode["title"],
        title_en: episode["titleEn"]
      )
    end

    def build_record(record)
      RecordEntity.new(
        database_id: record["databaseId"]
      )
    end
  end
end
