# frozen_string_literal: true
# == Schema Information
#
# Table name: likes
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  resource_id   :integer          not null
#  resource_type :string(510)      not null
#  created_at     :datetime
#  updated_at     :datetime
#
# Indexes
#
#  likes_user_id_idx  (user_id)
#

module Api
  module Internal
    class ReactionsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def add
        resource = params[:resource_type].constantize.find(params[:resource_id])
        current_user.add_reaction!(resource, params[:kind].to_sym)
        ga_client.page_category = params[:page_category]
        ga_client.events.create(:reactions, :create, el: recipient_type, ev: params[:kind], ds: "internal_api")

        head 201
      end

      def remove
        resource = params[:resource_type].constantize.find(params[:resource_id])
        current_user.remove_reaction!(resource, params[:kind].to_sym)

        head 200
      end
    end
  end
end
