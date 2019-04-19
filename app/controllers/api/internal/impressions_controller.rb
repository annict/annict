# frozen_string_literal: true

module Api
  module Internal
    class ImpressionsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def show
        work = Work.find(params[:work_id])
        @tags = current_user.tags_by_work(work)
        @all_tags = current_user.work_tags.published
        @popular_tags = WorkTag.published.popular_tags(work).limit(10)
        @comment = current_user.comment_by_work(work)
      end

      def update
        work = Work.find(params[:work_id])

        ActiveRecord::Base.transaction do
          current_user.update_work_tags!(work, params[:tags].presence || [])

          work_comment = current_user.work_comments.find_by(work: work)
          work_comment = current_user.work_comments.new(work: work) if work_comment.blank?

          if params[:comment].blank? || params[:comment] != work_comment.body
            work_comment.body = params[:comment].presence || ""
            work_comment.save!
          end
        end

        @tags = current_user.tags_by_work(work)
        @comment = current_user.comment_by_work(work)
      end
    end
  end
end
