# typed: false
# frozen_string_literal: true

module Annict
  module DataCare
    class MergeEpisode
      attr_reader :base_episode_id, :episode_id

      def initialize(base_episode_id, episode_id)
        @base_episode_id = base_episode_id
        @episode_id = episode_id
      end

      def run!
        ActiveRecord::Base.transaction do
          puts "Running merge_records! ..."
          merge_records!
          puts "Running merge_activities! ..."
          merge_activities!
          puts "Running merge_slots! ..."
          merge_slots!
          puts "Running destroy_episode! ..."
          destroy_episode!
        end
      end

      private

      def base_episode
        @base_episode ||= Episode.find(base_episode_id)
      end

      def episode
        @episode ||= Episode.find(episode_id)
      end

      def merge_records!
        episode.episode_records.find_each do |er|
          er.update(episode_id: base_episode.id, work_id: base_episode.work_id)
          er.record.update(work_id: base_episode.work_id)
          er.activities.update_all(
            trackable_id: er.id,
            recipient_id: base_episode_id,
            episode_id: base_episode.id,
            work_id: base_episode.work_id
          )
          er.likes.update_all(recipient_id: er.id)
          er.comments.update_all(episode_record_id: er.id)
          er.destroy_in_batches
        end
      end

      def merge_activities!
        episode.activities.update_all(
          recipient_id: base_episode_id,
          episode_id: base_episode.id,
          work_id: base_episode.work_id
        )
      end

      def merge_slots!
        episode.slots.update_all(episode_id: base_episode_id)
      end

      def destroy_episode!
        episode.destroy_in_batches
      end
    end
  end
end
