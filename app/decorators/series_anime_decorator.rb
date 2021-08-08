# frozen_string_literal: true

module SeriesAnimeDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || id
    link_to(name, db_edit_series_work_path(self), options)
  end

  def to_values
    self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :work_id
        path = anime_path(anime_id: anime.id)
        link_to(anime.local_title, path, target: "_blank")
      else
        send(field)
      end

      hash
    end
  end
end
