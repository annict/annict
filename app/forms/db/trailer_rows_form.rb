# frozen_string_literal: true

module Db
  class TrailerRowsForm
    include ActiveModel::Model
    include Virtus.model
    include ResourceRows

    row_model Trailer

    attr_accessor :work

    private

    def attrs_list
      trailers_count = @work.trailers.count
      @attrs_list ||= fetched_rows.map.with_index { |row_data, i|
        {
          work_id: @work.id,
          url: row_data[:url][:value],
          title: row_data[:title][:value],
          sort_number: (i + trailers_count) * 10
        }
      }
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        {
          url: {value: row_columns[0]},
          title: {value: row_columns[1]}
        }
      end
    end
  end
end
