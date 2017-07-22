# frozen_string_literal: true

module Api
  module Internal
    class CollectionItemsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create(collection_id, work_id, title, comment)
        @collection = current_user.collections.published.find(collection_id)
        collection_item = @collection.collection_items.new do |i|
          i.user_id = current_user.id
          i.work_id = work_id
          i.title = title
          i.comment = comment
        end

        return if collection_item.save

        render status: 400, json: { message: collection_item.errors.full_messages.first }
      end
    end
  end
end
