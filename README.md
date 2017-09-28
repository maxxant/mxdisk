# mxdisk

disks monitoring for insert, remove, mount, umount events

alternatives:

- in newers linux distributives: **findmnt -poll** for mounts events and **udevadm monitor** for disks and partitions events
- dbus + hal/udisk/udisk2

linux kernel "removable" disk property means physical layer and it is equal to external or hotplug device, for example linux/drivers/ata/libata-scsi.c:

```	/* set scsi removable (RMB) bit per ata bit, or if the
	 * AHCI port says it's external (Hotplug-capable, eSATA).
	 */
	if (ata_id_removable(args->id) ||
	    (args->dev->link->ap->pflags & ATA_PFLAG_EXTERNAL))
		    hdr[1] |= (1 << 7); ```
