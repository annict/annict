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

      def create
        recipient = params[:recipient_type].constantize.find(params[:recipient_id])
        current_user.like(recipient)
        ga_client.page_category = params[:page_category]
        ga_client.events.create(:likes, :create, el: params[:recipient_type], ev: params[:recipient_id], ds: "internal_api")
        keen_client.publish(:like_create, via: "internal_api", resource_type: params[:recipient_type])

        if params[:recipient_type] == "EpisodeRecord"
          EmailNotificationService.send_email(
            "liked_episode_record",
            recipient.user,
            current_user.id,
            params[:recipient_id]
          )
        end

        head 200
      end

      def unlike
        recipient = params[:recipient_type].constantize.find(params[:recipient_id])
        current_user.unlike(recipient)
        head 200
      end
    end
  end
end
