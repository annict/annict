# frozen_string_literal: true

module Types
  module Objects
    class UserType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

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
      field :created_at, Types::Scalars::DateTime, null: false
      field :viewer_can_follow, Boolean, null: false
      field :viewer_is_following, Boolean, null: false
      field :email, String, null: true
      field :notifications_count, Integer, null: true
      field :following, Types::Objects::UserType.connection_type, null: true
      field :followers, Types::Objects::UserType.connection_type, null: true

      field :activities, Connections::ActivityConnection, null: true, connection: true do
        argument :order_by, Types::InputObjects::ActivityOrder, required: false
      end

      field :following_activities, Connections::ActivityConnection, null: true, connection: true do
        argument :order_by, Types::InputObjects::ActivityOrder, required: false
      end

      field :records, Types::Objects::RecordType.connection_type, null: true do
        argument :order_by, Types::InputObjects::RecordOrder, required: false
        argument :has_comment, Boolean, required: false
      end

      field :works, Types::Objects::WorkType.connection_type, null: true do
        argument :annict_ids, [Integer], required: false
        argument :seasons, [String], required: false
        argument :titles, [String], required: false
        argument :state, Types::Enums::StatusState, required: false
        argument :order_by, Types::InputObjects::WorkOrder, required: false
      end

      field :programs, Types::Objects::ProgramType.connection_type, null: true do
        argument :unwatched, Boolean, required: false
        argument :order_by, Types::InputObjects::ProgramOrder, required: false
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
        ann_api_assets_url(object.profile, :image)
      end

      def background_image_url
        ann_api_assets_background_image_url(object.profile)
      end

      def records_count
        object.episode_records_count
      end

      def followings_count
        object.followings.without_deleted.count
      end

      def followers_count
        object.followers.without_deleted.count
      end

      def wanna_watch_count
        object.latest_statuses.count_on(:wanna_watch)
      end

      def watching_count
        object.latest_statuses.count_on(:watching)
      end

      def watched_count
        object.latest_statuses.count_on(:watched)
      end

      def on_hold_count
        object.latest_statuses.count_on(:on_hold)
      end

      def stop_watching_count
        object.latest_statuses.count_on(:stop_watching)
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
        ForeignKeyLoader.for(User, :id).load(object.followings.without_deleted.pluck(:id))
      end

      def followers
        ForeignKeyLoader.for(User, :id).load(object.followers.without_deleted.pluck(:id))
      end

      def activities(order_by: nil)
        SearchActivitiesQuery.new(
          object.activities,
          order_by: order_by
        ).call
      end

      def following_activities(order_by: nil)
        following_ids = object.followings.without_deleted.pluck(:id)
        following_ids << object.id
        SearchActivitiesQuery.new(
          Activity.where(user_id: following_ids),
          order_by: order_by
        ).call
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
          object.works.all,
          user: object,
          annict_ids: annict_ids,
          seasons: seasons,
          titles: titles,
          state: state,
          order_by: order_by
        ).call
      end

      def programs(unwatched: nil, order_by: nil)
        UserSlotsQuery.new(
          object,
          Slot.without_deleted.with_works(object.works_on(:wanna_watch, :watching).without_deleted),
          watched: unwatched.nil? ? nil : !unwatched,
          order: build_order(order_by)
        ).call
      end
    end
  end
end
