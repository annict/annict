module CastDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def to_values
      model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
        hash[field] = case field
        when :person_id
          person_id = send(:person_id)
          Person.find(person_id).name if person_id.present?
        when :work_id
          work_id = send(:work_id)
          Work.find(work_id).title if work_id.present?
        else
          send(field)
        end

        hash
      end
    end
  end
end
