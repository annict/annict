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
          puts "Running merge_acticities! ..."
          merge_acticities!
          puts "Running merge_slots! ..."
          merge_slots!
          puts "Running hide_episode! ..."
          hide_episode!
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
          attrs = er.attributes.except("id", "episode_id", "record_id")
          episode_record = base_episode.episode_records.create!(attrs) do |new_er|
            new_er.record = er.user.records.create!(work_id: er.work_id)
          end
          er.activities.update_all(trackable_id: episode_record.id, recipient_id: base_episode_id)
          er.likes.update_all(recipient_id: episode_record.id)
          er.comments.update_all(episode_record_id: episode_record.id)
          er.destroy_in_batches
        end
      end

      def merge_acticities!
        episode.activities.update_all(recipient_id: base_episode_id)
      end

      def merge_slots!
        episode.slots.update_all(episode_id: base_episode_id)
      end

      def hide_episode!
        episode.destroy_in_batches
      end
    end
  end
end
