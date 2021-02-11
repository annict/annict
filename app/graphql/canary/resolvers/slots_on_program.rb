# frozen_string_literal: true

module Canary
  module Resolvers
    class SlotsOnProgram < Canary::Resolvers::Base
      def resolve(viewer_untracked: nil, order_by: nil)
        viewer = context[:viewer]
        program = object
        order = Canary::OrderProperty.build(order_by)

        slot_ids = Rails.cache.fetch(cache_key(viewer, program, viewer_untracked), expires_in: 24.hours) do
          slots = program.slots.only_kept

          if viewer_untracked
            library_entry = viewer.library_entries.find_by(work_id: program.work_id)

            if library_entry
              slots = slots.where.not(episode_id: library_entry.watched_episode_ids)
            end
          end

          slots.pluck(:id)
        end

        Slot.where(id: slot_ids).order(order.field => order.direction)
      end

      private

      def cache_key(viewer, program, viewer_untracked)
        [
          self.class.name,
          viewer.id,
          program.id,
          program.updated_at.rfc3339,
          viewer_untracked.inspect,
          '2' # キャッシュクリア用の任意の文字列
        ].freeze
      end
    end
  end
end
