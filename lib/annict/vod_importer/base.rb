# frozen_string_literal: true

module Annict
  module VodImporter
    class Base
      def create_vod_title!(channel, attrs)
        vod_title_ids = []

        attrs.each do |attr|
          work = Work.published.find_by(title: attr[:name])

          print "name: #{attr[:name]} -> "

          conditions = if channel.amazon_video?
            { channel: channel, name: attr[:name] }
          else
            { channel: channel, code: attr[:code] }
          end
          vod_title = VodTitle.find_by(conditions)
          if vod_title
            puts "exists. id: #{vod_title.id}"
            vod_title_ids << nil
          else
            begin
              vod_title = VodTitle.create!(name: attr[:name], channel: channel, work: work, code: attr[:code])
              puts "created."
              vod_title_ids << vod_title.id
            rescue ActiveRecord::NotNullViolation => e
              Rails.logger.error(
                "[0d3248b1-e661-4032-8a53-9ad1515df8a9] create_vod_title! - " \
                "vod_title is not created. attr: #{attr}, message: #{e.message}"
              )
            end
          end

          next if work.blank?

          vod_title.hide! if vod_title.published?

          program_detail = work.program_details.published.find_by(channel: channel)
          if !program_detail
            work.program_details.create!(
              channel: channel,
              vod_title_code: vod_title.code,
              vod_title_name: vod_title.name
            )
          elsif program_detail.vod_title_code.blank?
            program_detail.update!(vod_title_code: vod_title.code, vod_title_name: vod_title.name)
          end
        end

        vod_title_ids
      end
    end
  end
end
