# frozen_string_literal: true

module Canary
  module Resolvers
    class Programs < Canary::Resolvers::Base
      def resolve(has_slots: nil, only_viewer_selected_channels: nil, order_by: nil)
        order = Canary::OrderProperty.build(order_by)

        programs = object.programs.only_kept

        if has_slots
          programs = programs.where.not(started_at: nil)
        end

        if only_viewer_selected_channels
          programs = programs.joins(:channel).merge(context[:viewer].channels.only_kept)
        end

        programs.order(order.field => order.direction)
      end
    end
  end
end
