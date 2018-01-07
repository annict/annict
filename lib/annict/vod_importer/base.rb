# frozen_string_literal: true

module Annict
  module VodImporter
    class Base
      def create_vod_title!(channel, attrs)
        vod_title_ids = []

        attrs.each do |attr|
          work = Work.published.find_by(title: attr[:name])

          print "name: #{attr[:name]} -> "

          vod_title = VodTitle.find_by(channel: channel, code: attr[:code])
          if vod_title.present?
            puts "exists. id: #{vod_title.id}"
            vod_title_ids << nil
          else
            vod_title = VodTitle.create!(name: attr[:name], channel: channel, work: work, code: attr[:code])
            puts "created."
            vod_title_ids << vod_title.id
          end

          next if work.blank?

          vod_title.hide! if vod_title.published?

          program_detail = work.program_details.published.find_by(channel: channel)
          next if program_detail.present?

          work.program_details.create!(channel: channel, vod_title_code: vod_title.code, vod_title_name: vod_title.name)
        end

        vod_title_ids
      end
    end
  end
end
