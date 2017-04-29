# frozen_string_literal: true

module Api
  module Internal
    class MultipleRecordsController < Api::Internal::ApplicationController
      permits :episode_ids

      before_action :authenticate_user!

      def create(episode_ids, page_category)
        records = MultipleRecordsService.new(current_user)
        records.save!(episode_ids)
        keen_client.page_category = page_category
        keen_client.multiple_records.create
        ga_client.page_category = page_category
        ga_client.events.create(:multiple_records, :create)
        flash[:notice] = t "messages.multiple_records.create.saved"
        head 201
      end
    end
  end
end
