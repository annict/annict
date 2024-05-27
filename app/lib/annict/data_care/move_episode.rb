# typed: false
# frozen_string_literal: true

module Annict
  module DataCare
    class MoveEpisode
      def initialize(episode_id, work_id)
        @episode_id = episode_id
        @work_id = work_id
      end

      def run!
        puts "Running move_records! ..."
        move_records!
        puts "Running move_acticities! ..."
        move_acticities!
        puts "Running move_slots! ..."
        move_slots!
        puts "Running move_episode! ..."
        move_episode!
      end

      private

      def episode
        @episode ||= Episode.find(@episode_id)
      end

      def move_records!
        episode.records.update_all(work_id: @work_id)
      end

      def move_acticities!
        episode.activities.update_all(work_id: @work_id)
      end

      def move_slots!
        episode.slots.update_all(work_id: @work_id)
      end

      def move_episode!
        episode.update_column(:work_id, @work_id)
      end
    end
  end
end
