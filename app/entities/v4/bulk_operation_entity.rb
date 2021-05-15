# frozen_string_literal: true

module V4
  class BulkOperationEntity < V4::ApplicationEntity
    attribute? :job_id, Types::String

    def self.from_nodes(nodes)
      nodes.map do |node|
        from_node(node)
      end
    end

    def self.from_node(node)
      attrs = {}

      if (job_id = node["jobId"])
        attrs[:job_id] = job_id
      end

      new attrs
    end
  end
end
