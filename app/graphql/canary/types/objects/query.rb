# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class Query < Canary::Types::Objects::Base
        add_field GraphQL::Types::Relay::NodeField
        add_field GraphQL::Types::Relay::NodesField

        field :viewer, Canary::Types::Objects::UserType,
          null: true,
          description: "認証されているユーザ"

        field :user, Canary::Types::Objects::UserType, null: true do
          argument :username, String, required: true
        end

        field :work_list, Canary::Types::Objects::WorkType.connection_type, null: true, resolver: Canary::Resolvers::WorkList do
          argument :database_ids, [Integer], required: false
          argument :seasons, [String], required: false
          argument :titles, [String], required: false
          argument :order_by, Canary::Types::InputObjects::WorkOrder, required: false
        end

        field :work, Canary::Types::Objects::WorkType, null: true do
          argument :database_id, Integer, required: true
        end

        field :episode, Canary::Types::Objects::EpisodeType, null: true do
          argument :database_id, Integer, required: true
        end

        field :people, Canary::Types::Objects::PersonType.connection_type, null: true do
          argument :database_ids, [Integer], required: false
          argument :names, [String], required: false
          argument :order_by, Canary::Types::InputObjects::PersonOrder, required: false
        end

        field :organizations, Canary::Types::Objects::OrganizationType.connection_type, null: true do
          argument :database_ids, [Integer], required: false
          argument :names, [String], required: false
          argument :order_by, Canary::Types::InputObjects::OrganizationOrder, required: false
        end

        field :characters, Canary::Types::Objects::CharacterType.connection_type, null: true do
          argument :database_ids, [Integer], required: false
          argument :names, [String], required: false
          argument :order_by, Canary::Types::InputObjects::CharacterOrder, required: false
        end

        field :channels, Canary::Types::Objects::ChannelType.connection_type, null: true do
          argument :is_vod, Boolean, required: false
        end

        field :activity_groups, Canary::Types::Objects::ActivityGroupType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::ActivityOrder, required: false
        end

        field :bulk_operation, Canary::Types::Objects::BulkOperationType, null: true do
          argument :job_id, String, required: true
        end

        def viewer
          context[:viewer]
        end

        def user(username:)
          User.only_kept.find_by(username: username)
        end

        def work(database_id:)
          Work.only_kept.find_by(id: database_id)
        end

        def episode(database_id:)
          Episode.only_kept.find_by(id: database_id)
        end

        def people(database_ids: nil, names: nil, order_by: nil)
          Deprecated::SearchPeopleQuery.new(
            annict_ids: database_ids,
            names: names,
            order_by: order_by
          ).call
        end

        def organizations(database_ids: nil, names: nil, order_by: nil)
          Deprecated::SearchOrganizationsQuery.new(
            annict_ids: database_ids,
            names: names,
            order_by: order_by
          ).call
        end

        def characters(database_ids: nil, names: nil, order_by: nil)
          Deprecated::SearchCharactersQuery.new(
            annict_ids: database_ids,
            names: names,
            order_by: order_by
          ).call
        end

        def channels(is_vod: nil)
          Deprecated::ChannelsQuery.new(
            Channel.only_kept,
            is_vod: is_vod
          ).call
        end

        def activity_groups(order_by: nil)
          order = Canary::OrderProperty.build(order_by)
          ActivityGroup.joins(:user).merge(User.only_kept).order(order.field => order.direction)
        end

        def bulk_operation(job_id:)
          job = Delayed::Job.find_by(id: job_id)
          job ? OpenStruct.new(job_id: job.id) : nil
        end
      end
    end
  end
end
