# typed: false
# frozen_string_literal: true

module Deprecated::Db
  class ProgramRowsForm
    include ActiveModel::Model
    T.unsafe(self).include Virtus.model
    include ResourceRows

    row_model Program

    attr_accessor :work

    validate :valid_time

    private

    def attrs_list
      @attrs_list ||= fetched_rows.map { |row_data|
        {
          channel_id: row_data[:channel][:id],
          work_id: @work.id,
          started_at: row_data[:started_at][:value],
          rebroadcast: row_data[:rebroadcast][:value],
          vod_title_code: row_data[:vod_title_code][:value],
          vod_title_name: row_data[:vod_title_name][:value],
          time_zone: @user.time_zone
        }
      }
    end

    def valid_time
      fetched_rows.each do |row_data|
        started_at = row_data[:started_at][:value]

        return true if started_at.blank?

        begin
          Time.parse(started_at)
        rescue
          i18n_path = "activemodel.errors.forms.db/slot_rows_form.invalid_start_time"
          errors.add(:rows, I18n.t(i18n_path))
        end
      end
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        channel = Channel.only_kept.where(id: row_columns[0])
          .or(Channel.only_kept.where(name: row_columns[0])).first

        {
          channel: {id: channel&.id, value: row_columns[0]},
          started_at: {value: row_columns[1]},
          rebroadcast: {value: row_columns[2] == "1"},
          vod_title_code: {value: row_columns[3].presence || ""},
          vod_title_name: {value: row_columns[4].presence || ""}
        }
      end
    end
  end
end
