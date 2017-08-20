# frozen_string_literal: true

module DB
  class ProgramDetailRowsForm
    include ActiveModel::Model
    include Virtus.model
    include ResourceRows

    row_model ProgramDetail

    attr_accessor :work

    validate :valid_time

    private

    def attrs_list
      @attrs_list ||= fetched_rows.map do |row_data|
        {
          channel_id: row_data[:channel][:id],
          work_id: @work.id,
          started_at: row_data[:started_at][:value],
          url: row_data[:url][:value],
          time_zone: @user.time_zone
        }
      end
    end

    def valid_time
      fetched_rows.each do |row_data|
        started_at = row_data[:started_at][:value]

        begin
          Time.parse(started_at)
        rescue
          i18n_path = "activemodel.errors.forms.db/program_rows_form.invalid_start_time"
          errors.add(:rows, I18n.t(i18n_path))
        end
      end
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        channel = Channel.published.where(id: row_columns[0]).
          or(Channel.published.where(name: row_columns[0])).first

        {
          channel: { id: channel&.id, value: row_columns[0] },
          started_at: { value: row_columns[1] },
          url: { value: row_columns[2] }
        }
      end
    end
  end
end
