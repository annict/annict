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
        puts "Running merge_latest_statuses! ..."
        merge_latest_statuses!
        puts "Running hide_work! ..."
        hide_work!
      end

      private

      def base_work
        @base_work ||= Work.find(base_work_id)
      end

      def work
        @work ||= Work.find(work_id)
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

      def merge_latest_statuses!
        work.latest_statuses.find_each do |ls|
          lstatus = ls.user.latest_statuses.where(work_id: base_work_id).first
          next if lstatus.present?
          attrs = ls.attributes.except("id")
          base_work.latest_statuses.create(attrs)
        end
      end

      def hide_work!
        work.soft_delete
      end
    end
  end
end
