# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class UserType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :username, String, null: false
        field :name, String, null: false
        field :description, String, null: false
        field :url, String, null: true
        field :avatar_url, String, null: true
        field :background_image_url, String, null: true
        field :records_count, Integer, null: false
        field :followings_count, Integer, null: false
        field :followers_count, Integer, null: false
        field :wanna_watch_count, Integer, null: false
        field :watching_count, Integer, null: false
        field :watched_count, Integer, null: false
        field :on_hold_count, Integer, null: false
        field :stop_watching_count, Integer, null: false
        field :created_at, Beta::Types::Scalars::DateTime, null: false
        field :viewer_can_follow, Boolean, null: false
        field :viewer_is_following, Boolean, null: false
        field :email, String, null: true
        field :notifications_count, Integer, null: true
        field :following, Beta::Types::Objects::UserType.connection_type, null: true
        field :followers, Beta::Types::Objects::UserType.connection_type, null: true

        field :activities, Beta::Connections::ActivityConnection, null: true, connection: true do
          argument :order_by, Beta::Types::InputObjects::ActivityOrder, required: false
        end

        field :following_activities, Beta::Connections::ActivityConnection, null: true, connection: true do
          argument :order_by, Beta::Types::InputObjects::ActivityOrder, required: false
        end

        field :records, Beta::Types::Objects::RecordType.connection_type, null: true do
          argument :order_by, Beta::Types::InputObjects::RecordOrder, required: false
          argument :has_comment, Boolean, required: false
        end

        field :works, Beta::Types::Objects::WorkType.connection_type, null: true do
          argument :annict_ids, [Integer], required: false
          argument :seasons, [String], required: false
          argument :titles, [String], required: false
          argument :state, Beta::Types::Enums::StatusState, required: false
          argument :order_by, Beta::Types::InputObjects::WorkOrder, required: false
        end

        field :programs, Beta::Types::Objects::ProgramType.connection_type, null: true, resolver: Beta::Resolvers::Programs do
          argument :unwatched, Boolean, required: false
          argument :order_by, Beta::Types::InputObjects::ProgramOrder, required: false
        end

        field :library_entries, Beta::Types::Objects::LibraryEntryType.connection_type, null: true, resolver: Beta::Resolvers::LibraryEntries do
          argument :states, [Beta::Types::Enums::StatusState], "視聴ステータス", required: false
          argument :seasons, [String], "指定したシーズンの作品を取得する", required: false
          argument :season_from, String, "指定したシーズンからの作品を取得する", required: false
          argument :season_until, String, "指定したシーズンまでの作品を取得する", required: false
          argument :order_by, Beta::Types::InputObjects::LibraryEntryOrder, required: false
        end

        def name
          object.profile.name
        end

        def description
          object.profile.description
        end

        def url
          object.profile.url
        end

        def avatar_url
          ann_avatar_image_url(object, width: 300, format: :jpg)
        end

        def background_image_url
          ann_image_url(object.profile, :background_image, width: 500, ratio: "16:9", format: :jpg)
        end

        def records_count
          object.episode_records_count
        end

        def followings_count
          object.followings.only_kept.count
        end

        def followers_count
          object.followers.only_kept.count
        end

        def wanna_watch_count
          object.library_entries.count_on(:wanna_watch)
        end

        def watching_count
          object.library_entries.count_on(:watching)
        end

        def watched_count
          object.library_entries.count_on(:watched)
        end

        def on_hold_count
          object.library_entries.count_on(:on_hold)
        end

        def stop_watching_count
          object.library_entries.count_on(:stop_watching)
        end

        def viewer_can_follow
          viewer = context[:doorkeeper_token].owner
          viewer != object && !viewer.following?(object)
        end

        def viewer_is_following
          viewer = context[:doorkeeper_token].owner
          viewer.following?(object)
        end

        def email
          return if context[:doorkeeper_token].owner != object

          object.email
        end

        def notifications_count
          return if context[:doorkeeper_token].owner != object

          object.notifications_count
        end

        def following
          Beta::ForeignKeyLoader.for(User, :id).load(object.followings.only_kept.pluck(:id))
        end

        def followers
          Beta::ForeignKeyLoader.for(User, :id).load(object.followers.only_kept.pluck(:id))
        end

        def activities(order_by: nil)
          order = build_order(order_by)
          object.activities.order(order.field => order.direction)
        end

        def following_activities(order_by: nil)
          object.following_resources(model: Activity, viewer: context[:viewer], order: build_order(order_by))
        end

        def records(order_by: nil, has_comment: nil)
          SearchEpisodeRecordsQuery.new(
            object.episode_records,
            order_by: order_by,
            has_body: has_comment
          ).call
        end

        def works(annict_ids: nil, seasons: nil, titles: nil, state: nil, order_by: nil)
          SearchWorksQuery.new(
            object.works,
            user: object,
            annict_ids: annict_ids,
            seasons: seasons,
            titles: titles,
            state: state,
            order_by: order_by
          ).call
        end
      end
    end
  end
end
