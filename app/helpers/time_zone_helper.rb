# frozen_string_literal: true

module TimeZoneHelper
  def display_time(time)
    time_zone = current_user&.time_zone.presence || "UTC"
    time&.in_time_zone(time_zone)&.strftime("%Y-%m-%d %H:%M")
  end

  def decorated_tz_name(time_zone)
    offset = time_zone.utc_offset / 3600
    formatted_offset = format("%02d:00", offset.abs)
    formatted_offset = offset >= 0 ? "+#{formatted_offset}" : "-#{formatted_offset}"
    "(GMT#{formatted_offset}) #{time_zone.name}"
  end

  def local_time_zone
    current_user&.time_zone.presence || cookies["ann_time_zone"]
  end
end
