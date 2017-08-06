# frozen_string_literal: true

module Api
  module Internal
    class ImpressionsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def show(work_id)
        work = Work.find(work_id)
        @tags = current_user.tags_by_work(work)
        @all_tags = current_user.work_tags.published
        @comment = current_user.comment_by_work(work)
      end

      def update(tags, comment, work_id)
        work = Work.find(work_id)

        ActiveRecord::Base.transaction do
          current_user.update_work_tag!(work, tags)

          if comment.present?
            work_comment = current_user.user_work_comments.find_by(work: work)
            work_comment = current_user.user_work_comments.new(work: work) if work_comment.blank?
            work_comment.body = comment
            work_comment.save!
          end
        end

        flash[:notice] = t("messages._common.updated")

        head 201
      end
    end
  end
end
