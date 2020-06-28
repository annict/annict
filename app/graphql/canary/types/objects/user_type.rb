# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class UserType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :database_id, Integer, null: false
        field :username, String, null: false
        field :name, String, null: false
        field :email, String, null: true
        field :description, String, null: false
        field :url, String, null: true
        field :locale, String, null: true
        field :avatar_url, String, null: true, description: "ユーザのアバター画像のURL" do
          argument :size, Canary::Types::Enums::UserAvatarSize, required: true
        end
        field :background_image_url, String, null: true
        field :viewer_can_follow, Boolean, null: false
        field :viewer_is_following, Boolean, null: false
        field :is_supporter, Boolean, null: false
        field :is_committer, Boolean, null: false
        field :locale, String, null: true
        field :display_supporter_badge, Boolean, null: false
        field :records_count, Integer, null: false
        field :following_count, Integer, null: false
        field :followers_count, Integer, null: false
        field :plan_to_watch_works_count, Integer, null: false
        field :watching_works_count, Integer, null: false
        field :completed_works_count, Integer, null: false
        field :on_hold_works_count, Integer, null: false
        field :dropped_works_count, Integer, null: false
        field :notifications_count, Integer, null: true
        field :character_favorites_count, Integer, null: false
        field :person_favorites_count, Integer, null: false
        field :organization_favorites_count, Integer, null: false
        field :created_at, Canary::Types::Scalars::DateTime, null: false
        field :following, Canary::Types::Objects::UserType.connection_type, null: true
        field :followers, Canary::Types::Objects::UserType.connection_type, null: true

        field :activity_groups, Canary::Types::Objects::ActivityGroupType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::ActivityOrder, required: false
        end

        field :activities, Canary::Types::Objects::ActivityType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::ActivityOrder, required: false
        end

        field :following_activity_groups, Canary::Types::Objects::ActivityGroupType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::ActivityOrder, required: false
        end

        field :following_activities, Canary::Types::Objects::ActivityType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::ActivityOrder, required: false
        end

        field :records, Canary::Types::Objects::RecordType.connection_type, null: true do
          argument :month, String, required: false
          argument :order_by, Canary::Types::InputObjects::RecordOrder, required: false
        end

        field :episode_records, Canary::Types::Objects::EpisodeRecordType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::EpisodeRecordOrder, required: false
          argument :has_body, Boolean, required: false
        end

        field :works, Canary::Types::Objects::WorkType.connection_type, null: true do
          argument :database_ids, [Integer], required: false
          argument :seasons, [String], required: false
          argument :titles, [String], required: false
          argument :status_kind, Canary::Types::Enums::StatusKind, required: false
          argument :order_by, Canary::Types::InputObjects::WorkOrder, required: false
        end

        field :slots, Canary::Types::Objects::SlotType.connection_type, null: true do
          argument :unwatched, Boolean, required: false
          argument :order_by, Canary::Types::InputObjects::SlotOrder, required: false
        end

        field :character_favorites, Canary::Types::Objects::CharacterFavoriteType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::CharacterFavoriteOrder, required: false
        end

        field :cast_favorites, Canary::Types::Objects::PersonFavoriteType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::PersonFavoriteOrder, required: false
        end

        field :staff_favorites, Canary::Types::Objects::PersonFavoriteType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::PersonFavoriteOrder, required: false
        end

        field :organization_favorites, Canary::Types::Objects::OrganizationFavoriteType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::OrganizationFavoriteOrder, required: false
        end

        def name
          Canary::RecordLoader.for(Profile, column: :user_id).load(object.id).then(&:name)
        end

        def description
          Canary::RecordLoader.for(Profile, column: :user_id).load(object.id).then(&:description)
        end

        def url
          object.profile.url
        end

        def avatar_url(size:)
          Canary::RecordLoader.for(Profile, column: :user_id).load(object.id).then do |profile|
            api_user_avatar_url(profile, size)
          end
        end

        def background_image_url
          Canary::RecordLoader.for(Profile, column: :user_id).load(object.id).then do |profile|
            ann_api_assets_background_image_url(profile)
          end
        end

        def is_supporter
          Canary::RecordLoader.for(Setting, column: :user_id).load(object.id).then do |setting|
            context[:viewer] == object ? object.supporter? : object.supporter? && !setting.hide_supporter_badge?
          end
        end

        def is_committer
          object.committer?
        end

        def display_supporter_badge
          Canary::RecordLoader.for(Setting, column: :user_id).load(object.id).then do |setting|
            context[:viewer] == object ? !setting.hide_supporter_badge? : object.supporter? && !setting.hide_supporter_badge?
          end
        end

        def viewer_can_follow
          viewer = context[:viewer]
          viewer != object && !viewer.following?(object)
        end

        def viewer_is_following
          viewer = context[:viewer]
          viewer.following?(object)
        end

        def email
          return if context[:viewer] != object
          object.email
        end

        def notifications_count
          return if context[:viewer] != object
          object.notifications_count
        end

        def following
          ForeignKeyLoader.for(User, :id).load(object.followings.only_kept.pluck(:id))
        end

        def followers
          ForeignKeyLoader.for(User, :id).load(object.followers.only_kept.pluck(:id))
        end

        def activity_groups(order_by: nil)
          order = build_order(order_by)
          object.activity_groups.order(order.field => order.direction)
        end

        def activities(order_by: nil)
          order = build_order(order_by)
          object.activities.order(order.field => order.direction)
        end

        def following_activity_groups(order_by: nil)
          object.following_resources(model: ActivityGroup, viewer: context[:viewer], order: build_order(order_by))
        end

        def following_activities(order_by: nil)
          object.following_resources(model: Activity, viewer: context[:viewer], order: build_order(order_by))
        end

        def records(month: nil, order_by: nil)
          collection = object.records.only_kept

          if month
            if !%r{[0-9]{4}-[0-9]{2}}.match?(month)
              raise GraphQL::ExecutionError, "The `month` argument should be like `2020-03`"
            end

            Time.zone = object.time_zone
            y, m = month.split("-").map(&:to_i)
            collection = collection.by_month(m, year: y)
          end

          order = build_order(order_by)
          collection.order(order.field => order.direction)
        end

        def episode_records(order_by: nil, has_body: nil)
          SearchEpisodeRecordsQuery.new(
            object.episode_records,
            order_by: order_by,
            has_body: has_body
          ).call
        end

        def works(database_ids: nil, seasons: nil, titles: nil, status_kind: nil, order_by: nil)
          SearchWorksQuery.new(
            object.works.all,
            user: object,
            annict_ids: database_ids,
            seasons: seasons,
            titles: titles,
            state: status_kind,
            order_by: order_by
          ).call
        end

        def slots(watched: nil, order_by: nil)
          UserSlotsQuery.new(
            object,
            Slot.only_kept.with_works(object.works_on(:wanna_watch, :watching).only_kept),
            watched: watched,
            order: build_order(order_by)
          ).call
        end

        def character_favorites(order_by: nil)
          order = build_order(order_by)
          object.character_favorites.order(order.field => order.direction)
        end

        def cast_favorites(order_by: nil)
          order = build_order(order_by)
          object.person_favorites.with_cast.order(order.field => order.direction)
        end

        def staff_favorites(order_by: nil)
          order = build_order(order_by)
          object.person_favorites.with_staff.order(order.field => order.direction)
        end

        def organization_favorites(order_by: nil)
          order = build_order(order_by)
          object.organization_favorites.order(order.field => order.direction)
        end
      end
    end
  end
end
