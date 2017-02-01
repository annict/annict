# frozen_string_literal: true
# == Schema Information
#
# Table name: likes
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  recipient_id   :integer          not null
#  recipient_type :string(510)      not null
#  created_at     :datetime
#  updated_at     :datetime
#
# Indexes
#
#  likes_user_id_idx  (user_id)
#

module Api
  module Internal
    class LikesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create(recipient_type, recipient_id, page_category)
        recipient = recipient_type.constantize.find(recipient_id)
        current_user.like(recipient)
        keen_client.page_category = page_category
        keen_client.likes.create(resource_type: recipient_type)
        head 200
      end

      def unlike(recipient_type, recipient_id)
        recipient = recipient_type.constantize.find(recipient_id)
        current_user.unlike(recipient)
        head 200
      end
    end
  end
end
