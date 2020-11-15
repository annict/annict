# frozen_string_literal: true

module Canary
  module Mutations
    class CreateAnimeRecord < Canary::Mutations::Base
      argument :anime_id, ID,
        required: true
      argument :comment, String,
        required: false,
        description: "作品への感想"
      WorkRecord::RATING_FIELDS.each do |rating_field|
        argument rating_field.to_s.camelcase(:lower).to_sym, Canary::Types::Enums::Rating,
          required: false,
          description: "作品への評価"
      end
      argument :share_to_twitter, Boolean,
        required: false,
        description: "作品への記録をTwitterでシェアするかどうか"

      field :record, Canary::Types::Objects::RecordType, null: true

      def resolve(
        anime_id:,
        comment: nil,
        rating_overall: nil,
        rating_animation: nil,
        rating_music: nil,
        rating_story: nil,
        rating_character: nil,
        share_to_twitter: nil
      )
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        anime = Work.only_kept.find_by_graphql_id(anime_id)

        anime_record = viewer.work_records.new(
          body: comment,
          rating_overall_state: rating_overall&.downcase,
          rating_animation_state: rating_animation&.downcase,
          rating_music_state: rating_music&.downcase,
          rating_story_state: rating_story&.downcase,
          rating_character_state: rating_character&.downcase,
          share_to_twitter: share_to_twitter
        )
        anime_record.work = anime
        anime_record.detect_locale!(:body)

        ActiveRecord::Base.transaction do
          anime_record.record = viewer.records.create!(work: anime)

          unless anime_record.valid?
            raise GraphQL::ExecutionError, anime_record.errors.full_messages.join(", ")
          end

          anime_record.save!

          activity_group = viewer.create_or_last_activity_group!(anime_record)
          viewer.activities.create!(itemable: anime_record, activity_group: activity_group)

          viewer.update_share_record_setting(share_to_twitter)

          if viewer.share_record_to_twitter?
            viewer.share_work_record_to_twitter(anime_record)
          end
        end

        {
          record: anime_record.record
        }
      end
    end
  end
end
