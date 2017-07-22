# frozen_string_literal: true

module Api
  module Internal
    class CollectionsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def index(work_id)
        @collections = current_user.collections.published.order(updated_at: :desc)
        @work = Work.published.find(work_id)
      end

      def create(title, description, work_id, page_category)
        collection = current_user.collections.new(title: title, description: description)

        if collection.save
          ga_client.page_category = page_category
          ga_client.events.create(:collections, :create)
          CreateCollectionActivityJob.perform_later(current_user.id, collection.id)
          @collections = current_user.collections.published.published.order(updated_at: :desc)
          @work = Work.published.find(work_id)
        else
          render status: 400, json: { message: collection.errors.full_messages.first }
        end
      end
    end
  end
end
