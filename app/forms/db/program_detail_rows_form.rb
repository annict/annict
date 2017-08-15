# frozen_string_literal: true

module DB
  class ProgramDetailRowsForm
    include ActiveModel::Model
    include Virtus.model
    include ResourceRows

    row_model ProgramDetail

    attr_accessor :work

    private

    def attrs_list
      @attrs_list ||= fetched_rows.map do |row_data|
        {
          channel_id: row_data[:channel][:id],
          work_id: @work.id,
          unique_id: row_data[:unique_id][:value],
          locale: row_data[:locale][:value]
        }
      end
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        channel = Channel.published.where(id: row_columns[0]).
          or(Channel.published.where(name: row_columns[0])).first

        {
          channel: { id: channel&.id, value: row_columns[0] },
          unique_id: { value: row_columns[1] },
          locale: { value: row_columns[2] }
        }
      end
    end
  end
end
