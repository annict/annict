# frozen_string_literal: true

module Deprecated::Db
  class StaffRowsForm
    include ActiveModel::Model
    T.unsafe(self).include Virtus.model
    include ResourceRows

    row_model Staff

    attr_accessor :work

    validate :valid_resource

    private

    def attrs_list
      roles = %i[ja en]
        .map { |l| I18n.t("enumerize.staff.role", locale: l).invert }
        .inject(&:merge)

      staffs_count = @work.staffs.count
      @attrs_list ||= fetched_rows.map.with_index { |row_data, i|
        role = roles[row_data[:role]]

        {
          work_id: @work.id,
          resource_id: row_data[:resource][:id],
          resource_type: row_data[:resource][:type],
          role: (role.presence || :other),
          role_other: role.blank? ? row_data[:role] : nil,
          sort_number: (i + staffs_count) * 10
        }
      }
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        person = Person.only_kept.where(id: row_columns[1])
          .or(Person.only_kept.where(name: row_columns[1])).first
        organization = Organization.only_kept.where(id: row_columns[2])
          .or(Organization.only_kept.where(name: row_columns[2])).first

        resource, value = if person.present?
          [person, row_columns[1]]
        elsif organization.present?
          [organization, row_columns[2]]
        else
          [nil, row_columns[1].presence || row_columns[2]]
        end

        {
          role: row_columns[0],
          resource: {id: resource&.id, type: resource.class.name, value: value}
        }
      end
    end

    def valid_resource
      fetched_rows.each do |row_data|
        row_data.slice(:resource).each do |_, data|
          next if data[:id].present?
          i18n_path = "activemodel.errors.forms.db/staff_rows_form.invalid"
          errors.add(:rows, I18n.t(i18n_path, value: data[:value]))
        end
      end
    end
  end
end
