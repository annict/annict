describe Api::V1::WorksController do
  let(:token) { double(acceptable?: true) }

  before do
    allow(controller).to receive(:doorkeeper_token) { token }
  end

  describe "GET index" do
    it "200が返ること" do
      get :index
      expect(response.status).to eq(200)
    end
  end
end
