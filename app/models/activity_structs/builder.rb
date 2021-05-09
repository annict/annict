# frozen_string_literal: true

module ActivityStructs
  class Builder
    ACTIVITY_GROUP_MAPPING = {
      id: "activity_groups.id",
      itemable_type: "activity_groups.itemable_type",
      outstanding: "activity_groups.single",
      items_count: "activity_groups.activities_count",
      created_at: "activity_groups.created_at",
      user_id: "activity_groups.user_id"
    }.freeze
    USER_MAPPING = {
      id: "users.id",
      username: "users.username",
    }.freeze
    ACTIVITY_MAPPING = {
      activity_group_id: "activities.activity_group_id",
      trackable_type: "activities.trackable_type",
      trackable_id: "activities.trackable_id"
    }.freeze
    STATUS_MAPPING = {
      id: "statuses.id",
      kind: "statuses.kind",
      anime_id: "statuses.work_id",
      # anime_title: "works.title",
      # anime_title_en: "works.title_en"
    }
    ANIME_MAPPING = {
      id: "works.id",
      title: "works.title",
      title_en: "works.title_en",
  }


    def initialize(activity_groups:, current_user: nil)
      @activity_groups = activity_groups
      @current_user = current_user
    end

    def call
      all_activity_groups.map do |attrs|
        activities = all_activities.find_all { |a| a[:activity_group_id] == attrs[:id] }
        trackable_ids = activities.pluck(:trackable_id)
        statuses = all_statuses.find_all { |s| trackable_ids.include?(s[:id]) }

        attrs[:user] = all_users.find { |u| u[:id] == attrs[:user_id] }

        case attrs[:itemable_type]
        when "Status"
          attrs[:item_kind] = "status"
          attrs[:items] = statuses.map do |status|
            anime = all_anime_list.find { |a| a[:id] == status[:anime_id] }
            anime_image = all_anime_images.find { |ai| ai[:anime_id] == anime[:id] }

            ActivityStructs::StatusStruct.new(
              kind: Status.kind_v2_to_v3(status[:kind]).to_s,
              anime_id: anime[:id],
              anime: {
                title: anime[:title],
                title_en: anime[:title_en],
                image_path: anime_image&.uploaded_file_path(:image).presence || "no-image.jpg"
              }
            )
          end

          ActivityStructs::ActivityStruct.new(
            item_kind: attrs[:item_kind],
            outstanding: attrs[:outstanding],
            items_count: attrs[:items_count],
            created_at: attrs[:created_at],
            items: attrs[:items],
            user: {
              username: attrs[:user][:username]
            }
          )
        end
      end

      # all_activities = Activity.where(activity_group_id: activity_groups)
      # all_status_activities = all_activities.find_all { |activity| activity.trackable_type == "Status" }
      # all_episode_record_activities = all_activities.find_all { |activity| activity.trackable_type == "EpisodeRecord" }
      # all_anime_record_activities = all_activities.find_all { |activity| activity.trackable_type == "WorkRecord" }
      # all_statuses = Status
      #   .where(id: all_status_activities.pluck(:trackable_id))
      #   .preload(anime: :anime_image)
      #   .order(created_at: :desc)
      #   .pluck(:id, :kind)
      # all_records = Record.eager_load(:user, :work_record, episode_record: [:episode, {anime: :anime_image}])
      #   .merge(EpisodeRecord.where(id: all_episode_record_activities.pluck(:trackable_id)).or(
      #     AnimeRecord.where(id: all_anime_record_activities.pluck(:trackable_id))
      #   ))
      #   .order(created_at: :desc)
      # all_episode_records = EpisodeRecord.where(id: all_episode_record_activities.pluck(:trackable_id))
      # all_anime_records = AnimeRecord.where(id: all_anime_record_activities.pluck(:trackable_id))

      # activity_groups.map do |activity_group|
      #   trackable_ids = all_activities
      #     .pluck(:activity_group_id, :trackable_id)
      #     .find_all { |(activity_group_id, _)| activity_group_id == activity_group.id }
      #     .map { |(_, trackable_id)| trackable_id }

      #   itemable_type, itemables = case activity_group.itemable_type
      #   when "Status"
      #     statuses = all_statuses.find_all { |(status_id, _)| trackable_ids.include?(status_id) }.first(2)
      #     ["status", ]
      #   when "EpisodeRecord"
      #     episode_records = all_episode_records.find_all { |episode_record| trackable_ids.include?(episode_record.id) }
      #     ["record", all_records.find_all { |record| episode_records.pluck(:record_id).include?(record.id) }.first(2)]
      #   when "WorkRecord"
      #     anime_records = all_anime_records.find_all { |anime_record| trackable_ids.include?(anime_record.id) }
      #     ["record", all_records.find_all { |record| anime_records.pluck(:record_id).include?(record.id) }.first(2)]
      #   end

      #   Builder::Activity::ActivityGroupStruct.new(
      #     itemable_type: itemable_type,
      #     user: activity_group.user,
      #     created_at: activity_group.created_at,
      #     itemables: itemables,
      #     outstanding: activity_group.single?
      #   )
      # end
    end

    private

    def all_activity_groups
      @all_activity_groups ||= begin
        if @current_user
          @activity_groups
            .pluck(*ACTIVITY_GROUP_MAPPING.values)
            .map { |ag| ACTIVITY_GROUP_MAPPING.keys.zip(ag).to_h }
        end
      end
    end

    def all_users
      @all_users ||= User
        .where(id: all_activity_groups.pluck(:user_id))
        .pluck(*USER_MAPPING.values)
        .map { |u| USER_MAPPING.keys.zip(u).to_h }
    end

    def all_profiles
      @all_profiles ||= Profile
        .where(user_id: all_users.pluck(:id))
        .select(:image_data)
    end

    def all_activities
      @all_activities ||= Activity
        .where(activity_group_id: all_activity_groups.pluck(:id))
        .pluck(*ACTIVITY_MAPPING.values)
        .map { |a| ACTIVITY_MAPPING.keys.zip(a).to_h }
      end

    def all_statuses
      @all_statuses ||= Status
        .where(id: all_activities.find_all { |a| a[:trackable_type] == "Status" }.pluck(:trackable_id))
        .pluck(*STATUS_MAPPING.values)
        .map { |s| STATUS_MAPPING.keys.zip(s).to_h }
    end

    def all_anime_list
      @all_anime_list ||= Anime
        .where(id: all_statuses.pluck(:anime_id))
        .pluck(*ANIME_MAPPING.values)
        .map { |a| ANIME_MAPPING.keys.zip(a).to_h }
    end

    def all_anime_images
      @all_anime_images ||= WorkImage
        .where(work_id: all_statuses.pluck(:anime_id))
        .select(:image_data)
    end
  end
end
