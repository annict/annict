# frozen_string_literal: true

module Api
  module Internal
    class CollectionsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def index(work_id)
        @collections = current_user.collections.published.order(updated_at: :desc)
        @work = Work.published.find(work_id)
      end

      def create(title, description, work_id)
        collection = current_user.collections.new(title: title, description: description)

        if collection.save
          @collections = current_user.collections.published.published.order(updated_at: :desc)
          @work = Work.published.find(work_id)
        else
          render status: 400, json: { message: collection.errors.full_messages.first }
        end
      end
    end
  end
end
