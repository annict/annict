# frozen_string_literal: true

module SeriesWorkDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || id
    h.link_to(name, h.edit_db_series_work_path(self), options)
  end

  def to_values
    model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :work_id
        path = h.work_path(work)
        h.link_to(work.decorate.local_title, path, target: "_blank")
      else
        send(field)
      end

      hash
    end
  end
end
