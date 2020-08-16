# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class RecordType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :database_id, Integer, null: false
        field :itemable_type, Canary::Types::Enums::RecordItemableType, null: false
        field :modified_at, Canary::Types::Scalars::DateTime, null: true
        field :created_at, Canary::Types::Scalars::DateTime, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :anime, Canary::Types::Objects::AnimeType, null: false
        field :itemable, Canary::Types::Unions::RecordItemable, null: false

        def itemable_type
          Canary::AssociationLoader.for(Record, %i(episode_record)).load(object).then do |episode_record|
            episode_record.first ? "EPISODE_RECORD" : "ANIME_RECORD"
          end
        end

        def modified_at
          itemable_type.then do |i_type|
            case i_type
            when "EPISODE_RECORD"
              Canary::AssociationLoader.for(Record, %i(episode_record)).load(object).then do |episode_record|
                episode_record.first.modify_body ? episode_record.first.updated_at : nil
              end
            when "ANIME_RECORD"
              Canary::AssociationLoader.for(Record, %i(work_record)).load(object).then do |work_record|
                work_record.first.modified_at
              end
            end
          end
        end

        def user
          Canary::RecordLoader.for(User).load(object.user_id)
        end

        def anime
          Canary::RecordLoader.for(Work).load(object.work_id)
        end

        def itemable
          itemable_type.then do |i_type|
            case i_type
            when "EPISODE_RECORD"
              Canary::AssociationLoader.for(Record, %i(episode_record)).load(object).then(&:first)
            when "ANIME_RECORD"
              Canary::AssociationLoader.for(Record, %i(work_record)).load(object).then(&:first)
            end
          end
        end
      end
    end
  end
end
