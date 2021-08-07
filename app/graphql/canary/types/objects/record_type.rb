# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class RecordType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node
        implements Canary::Types::Interfaces::Reactable

        global_id_field :id

        field :database_id, Integer, null: false
        field :comment, String, null: true
        field :comment_html, String, null: true
        field :likes_count, Integer, null: false
        field :modified_at, Canary::Types::Scalars::DateTime, null: true
        field :created_at, Canary::Types::Scalars::DateTime, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :trackable, Canary::Types::Unions::RecordTrackable, null: false
        field :recordable, Canary::Types::Unions::Recordable, null: false

        def comment
          recordable_type.then do |type|
            case type
            when "EpisodeRecord"
              Canary::AssociationLoader.for(Record, %i[episode_record]).load(object).then do |episode_record|
                episode_record.first.body
              end
            when "AnimeRecord"
              Canary::AssociationLoader.for(Record, %i[anime_record]).load(object).then do |work_record|
                work_record.first.body
              end
            end
          end
        end

        def comment_html
          recordable_type.then do |type|
            case type
            when "EpisodeRecord"
              Canary::AssociationLoader.for(Record, %i[episode_record]).load(object).then do |episode_record|
                render_markdown(episode_record.first.body)
              end
            when "AnimeRecord"
              Canary::AssociationLoader.for(Record, %i[anime_record]).load(object).then do |work_record|
                render_markdown(work_record.first.body)
              end
            end
          end
        end

        def likes_count
          recordable.then(&:likes_count)
        end

        def modified_at
          recordable_type.then do |type|
            case type
            when "EpisodeRecord"
              Canary::AssociationLoader.for(Record, %i[episode_record]).load(object).then do |episode_record|
                episode_record.first.modify_body ? episode_record.first.updated_at : nil
              end
            when "AnimeRecord"
              Canary::AssociationLoader.for(Record, %i[anime_record]).load(object).then do |work_record|
                work_record.first.modified_at
              end
            end
          end
        end

        def user
          Canary::RecordLoader.for(User).load(object.user_id)
        end

        def trackable
          recordable.then do |comp|
            if comp.is_a?(EpisodeRecord)
              Canary::RecordLoader.for(Episode).load(comp.episode_id)
            else
              Canary::RecordLoader.for(Anime).load(comp.work_id)
            end
          end
        end

        def recordable
          recordable_type.then do |type|
            case type
            when "EpisodeRecord"
              Canary::AssociationLoader.for(Record, %i[episode_record]).load(object).then(&:first)
            when "AnimeRecord"
              Canary::AssociationLoader.for(Record, %i[anime_record]).load(object).then(&:first)
            end
          end
        end

        def reactions
          recordable.then(&:likes)
        end

        private

        def recordable_type
          Canary::AssociationLoader.for(Record, %i[episode_record]).load(object).then do |episode_record|
            episode_record.first ? "EpisodeRecord" : "AnimeRecord"
          end
        end
      end
    end
  end
end
