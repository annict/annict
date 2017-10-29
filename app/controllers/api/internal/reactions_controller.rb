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

      def add(resource_type, resource_id, kind, page_category)
        resource = resource_type.constantize.find(resource_id)
        current_user.add_reaction!(resource, kind.to_sym)
        ga_client.page_category = page_category
        ga_client.events.create(:reactions, :create)
        keen_client.publish(
          "create_reactions",
          user: current_user,
          page_category: page_category,
          via: "internal_api",
          resource_type: recipient_type,
          kind: kind
        )

        head 201
      end

      def remove(resource_type, resource_id, kind)
        resource = resource_type.constantize.find(resource_id)
        current_user.remove_reaction!(resource, kind.to_sym)

        head 200
      end
    end
  end
end
