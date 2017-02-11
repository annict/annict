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
        puts "Running merge_records! ..."
        merge_records!
        puts "Running merge_acticities! ..."
        merge_acticities!
        puts "Running merge_programs! ..."
        merge_programs!
        puts "Running hide_episode! ..."
        hide_episode!
      end

      private

      def base_episode
        @base_episode ||= Episode.find(base_episode_id)
      end

      def episode
        @episode ||= Episode.find(episode_id)
      end

      def merge_records!
        episode.records.find_each do |c|
          attrs = c.attributes.except("id")
          checkin = base_episode.records.create(attrs)
          c.activities.update_all(trackable_id: checkin.id, recipient_id: base_episode_id)
          c.likes.update_all(recipient_id: checkin.id)
          c.comments.update_all(checkin_id: checkin.id)
        end
      end

      def merge_acticities!
        episode.activities.update_all(recipient_id: base_episode_id)
      end

      def merge_programs!
        episode.programs.update_all(episode_id: base_episode_id)
      end

      def hide_episode!
        episode.update_column(:aasm_state, "hidden")
      end
    end
  end
end
