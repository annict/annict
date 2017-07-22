# frozen_string_literal: true

class CreateCollectionActivityJob < ApplicationJob
  queue_as :default

  def perform(user_id, collection_id)
    user = User.find(user_id)
    collection = user.collections.published.find(collection_id)

    Activity.create! do |a|
      a.user = user
      a.recipient = user
      a.trackable = collection
      a.action = "create_collection"
    end
  end
end
