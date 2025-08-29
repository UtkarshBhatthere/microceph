.. _howto_active_active_nfs:

How to create an active-active NFS Ganesha service
===================================================

This guide explains how to create a highly available, active-active NFS Ganesha service using the MicroCeph ingress feature. The ingress service uses keepalived and haproxy to provide a virtual IP (VIP) that load balances traffic across multiple NFS Ganesha instances. You can also use this feature to provide ingress for other services.

Prerequisites
-------------

- A running MicroCeph cluster with at least two nodes.
- The ``nfs`` feature enabled on at least two nodes, each belonging to the same NFS cluster.

1. Enable NFS Ganesha services
-------------------------------

First, enable the NFS Ganesha service on two or more nodes. Make sure to use the same ``--cluster-id`` for each instance to group them into a single NFS cluster.

On the first node:

.. code-block:: bash

   microceph enable nfs --cluster-id nfs-ha --target-node node1

On the second node:

.. code-block:: bash

   microceph enable nfs --cluster-id nfs-ha --target-node node2

This will create two NFS Ganesha instances that are part of the ``nfs-ha`` cluster.

2. Enable the ingress service
-----------------------------

Next, enable the ingress service. This will create a VIP that floats between the nodes where the ingress service is enabled and load balances traffic to the target service instances.

.. code-block:: bash

   microceph enable ingress --service-id ingress-nfs-ha \
     --vip-address 192.168.1.100 \
     --vip-interface eth0 \
     --target nfs.nfs-ha \
     --target-node node1

Repeat the same command for any other node where you want to run the ingress service (typically the same nodes as your target service).

.. code-block:: bash

   microceph enable ingress --service-id ingress-nfs-ha \
     --vip-address 192.168.1.100 \
     --vip-interface eth0 \
     --target nfs.nfs-ha \
     --target-node node2

- ``--service-id``: A unique name for this ingress service instance.
- ``--vip-address``: The virtual IP address that clients will connect to.
- ``--vip-interface``: The network interface on which the VIP will be active.
- ``--target``: The service to provide ingress for, in the format ``<service-name>.<service-id>``. In this case, we are targeting the ``nfs`` service with the ID ``nfs-ha``.

MicroCeph will automatically generate and manage the necessary VRRP password and router ID for the keepalived configuration.

Creating multiple ingress services
----------------------------------

You can run the ``enable ingress`` command multiple times with different ``--service-id`` values to create multiple, independent ingress services. This is useful for providing VIPs for different services or different clusters of the same service.

3. Verify the setup
-------------------

You should now be able to mount the NFS share using the virtual IP address:

.. code-block:: bash

   mount -t nfs -o port=2049 192.168.1.100:/ /mnt/nfs

4. Disable the ingress service
------------------------------

To disable an ingress service, use the ``disable ingress`` command with the service ID:

.. code-block:: bash

   microceph disable ingress --service-id ingress-nfs-ha --target-node node1
   microceph disable ingress --service-id ingress-nfs-ha --target-node node2

This will remove the configuration for this specific ingress service and reload the ingress service. If no other ingress services are configured, the `keepalived` and `haproxy` daemons will be stopped.
