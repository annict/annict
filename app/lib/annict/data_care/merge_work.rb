# frozen_string_literal: true

module Annict
  module DataCare
    class MergeWork
      attr_reader :base_work_id, :work_id

      def initialize(base_work_id, work_id)
        @base_work_id = base_work_id
        @work_id = work_id
      end

      def run!
        puts "Running merge_statuses! ..."
        merge_statuses!
        puts "Running merge_library_entries! ..."
        merge_library_entries!
        puts "Running hide_work! ..."
        hide_work!
      end

      private

      def base_work
        @base_work ||= Anime.find(base_work_id)
      end

      def work
        @work ||= Anime.find(work_id)
      end

      def merge_statuses!
        work.statuses.find_each do |s|
          status = s.user.statuses.where(work_id: base_work_id).first
          next if status.present?
          attrs = s.attributes.except("id")
          status = base_work.statuses.create(attrs)
          s.activities.update_all(trackable_id: status.id, recipient_id: base_work_id)
          s.likes.update_all(recipient_id: status.id)
        end
      end

      def merge_library_entries!
        work.library_entries.find_each do |ls|
          lstatus = ls.user.library_entries.where(work_id: base_work_id).first
          next if lstatus.present?
          attrs = ls.attributes.except("id")
          base_work.library_entries.create(attrs)
        end
      end

      def hide_work!
        work.destroy_in_batches
      end
    end
  end
end
