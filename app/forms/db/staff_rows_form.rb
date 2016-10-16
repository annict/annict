# frozen_string_literal: true

module DB
  class StaffRowsForm
    include ActiveModel::Model
    include Virtus.model
    include ResourceRows

    row_model Staff

    attr_accessor :work

    validate :valid_resource

    private

    def attrs_list
      roles = %i(ja en).
        map { |l| I18n.t("enumerize.staff.role", locale: l).invert }.
        inject(&:merge)

      @attrs_list ||= fetched_rows.map do |row_data|
        role = roles[row_data[:role]]

        {
          work_id: @work.id,
          resource_id: row_data[:resource][:id],
          resource_type: row_data[:resource][:type],
          role: (role.presence || :other),
          role_other: role.blank? ? row_data[:role] : nil
        }
      end
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        person = Person.where(id: row_columns[1]).
          or(Person.where(name: row_columns[1])).first
        organization = Organization.where(id: row_columns[2]).
          or(Organization.where(name: row_columns[2])).first

        resource, value = if person.present?
          [person, row_columns[1]]
        else
          [organization, (row_columns[2].presence || row_columns[1])]
        end

        {
          role: row_columns[0],
          resource: { id: resource&.id, type: resource.class.name, value: value }
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
