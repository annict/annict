# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class Query < Canary::Types::Objects::Base
        field :node, field: GraphQL::Relay::Node.field
        field :nodes, field: GraphQL::Relay::Node.plural_field
        field :viewer, Canary::Types::Objects::UserType,
          null: true,
          description: "認証されているユーザ"

        field :user, Canary::Types::Objects::UserType, null: true do
          argument :username, String, required: true
        end

        field :works, Canary::Types::Objects::WorkType.connection_type, null: true do
          argument :annict_ids, [Integer], required: false
          argument :seasons, [String], required: false
          argument :titles, [String], required: false
          argument :order_by, Canary::Types::InputObjects::WorkOrder, required: false
        end

        field :episodes, Canary::Types::Objects::EpisodeType.connection_type, null: true do
          argument :annict_ids, [Integer], required: false
          argument :order_by, Canary::Types::InputObjects::EpisodeOrder, required: false
        end

        field :people, Canary::Types::Objects::PersonType.connection_type, null: true do
          argument :annict_ids, [Integer], required: false
          argument :names, [String], required: false
          argument :order_by, Canary::Types::InputObjects::PersonOrder, required: false
        end

        field :organizations, Canary::Types::Objects::OrganizationType.connection_type, null: true do
          argument :annict_ids, [Integer], required: false
          argument :names, [String], required: false
          argument :order_by, Canary::Types::InputObjects::OrganizationOrder, required: false
        end

        field :characters, Canary::Types::Objects::CharacterType.connection_type, null: true do
          argument :annict_ids, [Integer], required: false
          argument :names, [String], required: false
          argument :order_by, Canary::Types::InputObjects::CharacterOrder, required: false
        end

        field :channels, Canary::Types::Objects::ChannelType.connection_type, null: true do
          argument :is_vod, Boolean, required: false
        end

        def viewer
          context[:viewer]
        end

        def user(username:)
          User.without_deleted.find_by(username: username)
        end

        def works(annict_ids: nil, seasons: nil, titles: nil, order_by: nil)
          SearchWorksQuery.new(
            annict_ids: annict_ids,
            seasons: seasons,
            titles: titles,
            order_by: order_by
          ).call
        end

        def episodes(annict_ids: nil, order_by: nil)
          SearchEpisodesQuery.new(
            annict_ids: annict_ids,
            order_by: order_by
          ).call
        end

        def people(annict_ids: nil, names: nil, order_by: nil)
          SearchPeopleQuery.new(
            annict_ids: annict_ids,
            names: names,
            order_by: order_by
          ).call
        end

        def organizations(annict_ids: nil, names: nil, order_by: nil)
          SearchOrganizationsQuery.new(
            annict_ids: annict_ids,
            names: names,
            order_by: order_by
          ).call
        end

        def characters(annict_ids: nil, names: nil, order_by: nil)
          SearchCharactersQuery.new(
            annict_ids: annict_ids,
            names: names,
            order_by: order_by
          ).call
        end

        def channels(is_vod: nil)
          SearchChannelsRepository.new(
            Channel.without_deleted,
            is_vod: is_vod
          ).call
        end
      end
    end
  end
end
