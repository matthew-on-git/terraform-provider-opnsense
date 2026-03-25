# Manage a firewall category for web traffic rules
resource "opnsense_firewall_category" "web_traffic" {
  name  = "web_traffic"
  color = "0000ff"
}

# Manage a category for management access rules
resource "opnsense_firewall_category" "management" {
  name  = "management"
  color = "ff0000"
}

# Manage a category with no color
resource "opnsense_firewall_category" "monitoring" {
  name = "monitoring"
}
