# frozen_string_literal: true

module SeriesWorkDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || id
    link_to(name, edit_db_series_work_path(self), options)
  end

  def to_values
    self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :work_id
        path = work_path(work)
        link_to(work.local_title, path, target: "_blank")
      else
        send(field)
      end

      hash
    end
  end
end
