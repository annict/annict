# frozen_string_literal: true

module Canary
  module Mutations
    class UpdateAnimeRecord < Canary::Mutations::Base
      argument :anime_record_id, ID, required: true
      argument :body, String,
        required: false,
        description: "作品への感想"
      WorkRecord::STATES.each do |state|
        argument state.to_s.camelcase(:lower).to_sym, Canary::Types::Enums::RatingState,
          required: true,
          description: "作品への評価"
      end
      argument :share_twitter, Boolean,
        required: false,
        description: "作品への記録をTwitterでシェアするかどうか"
      argument :share_facebook, Boolean,
        required: false,
        description: "作品への記録をFacebookでシェアするかどうか"

      field :anime_record, Canary::Types::Objects::AnimeRecordType, null: true

      def resolve(
        anime_record_id:,
        body: nil,
        rating_overall_state: nil,
        rating_animation_state: nil,
        rating_music_state: nil,
        rating_story_state: nil,
        rating_character_state: nil,
        share_twitter: nil,
        share_facebook: nil
      )
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        work_record = viewer.work_records.only_kept.find_by_graphql_id(anime_record_id)

        work_record.body = body
        WorkRecord::STATES.each do |state|
          work_record.send("#{state}=".to_sym, send(state.to_s.camelcase(:lower).to_sym)&.downcase)
        end
        work_record.modified_at = Time.zone.now
        work_record.oauth_application = context[:application]
        work_record.detect_locale!(:body)

        viewer.setting.attributes = {
          share_review_to_twitter: share_twitter == true,
          share_review_to_facebook: share_facebook == true
        }

        work_record.save!
        viewer.setting.save!
        viewer.touch(:record_cache_expired_at)

        if share_twitter
          viewer.share_work_record_to_twitter(work_record)
        end

        {
          anime_record: work_record
        }
      end
    end
  end
end
