# frozen_string_literal: true

describe "GraphQL Schema" do
  let(:dumped_schema_path) { Rails.root.join("app", "graphql", "beta", "schema.graphql") }
  let(:dumped_definition) { File.read(dumped_schema_path) }
  let(:current_definition) { Beta::AnnictSchema.to_definition }

  it "equals dumped schema" do
    expect(current_definition).to eq(dumped_definition)
  end
end
