# frozen_string_literal: true

module Canary
  module Mutations
    class CreateEpisodeRecord < Canary::Mutations::Base
      argument :episode_id, ID,
        required: true
      argument :comment, String,
        required: false,
        description: "エピソードへの感想"
      argument :rating, Canary::Types::Enums::RatingState,
        required: false,
        description: "エピソードへの評価"
      argument :share_to_twitter, Boolean,
        required: false,
        description: "エピソードへの記録をTwitterでシェアするかどうか"

      field :record, Canary::Types::Objects::RecordType, null: true

      def resolve(episode_id:, comment: nil, rating: nil, share_to_twitter: nil)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        episode = Episode.only_kept.find_by_graphql_id(episode_id)
        work = episode.work

        episode_record = viewer.episode_records.new(
          rating_state: rating&.downcase,
          body: comment,
          share_to_twitter: share_to_twitter
        )
        episode_record.episode = episode
        episode_record.work = work
        episode_record.detect_locale!(:body)

        library_entry = viewer.library_entries.find_by(work: work)

        ActiveRecord::Base.transaction do
          episode_record.record = viewer.records.create!(work: work)

          unless episode_record.valid?
            raise GraphQL::ExecutionError, episode_record.errors.full_messages.join(", ")
          end

          episode_record.save

          activity_group = viewer.create_or_last_activity_group!(episode_record)
          viewer.activities.create!(itemable: episode_record, activity_group: activity_group)

          viewer.update_share_record_setting(share_to_twitter)
          library_entry&.append_episode!(episode)

          if viewer.share_record_to_twitter?
            viewer.share_episode_record_to_twitter(episode_record)
          end
        end

        {
          record: episode_record.record
        }
      end
    end
  end
end
